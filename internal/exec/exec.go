package exec

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Binary is the CLI shipped inside the Obsidian app bundle. It must be the
// exact name `obsidian-cli`: the bundle's MacOS directory holds both the CLI
// and the app binary `Obsidian`, and on a case-insensitive filesystem a lookup
// for `obsidian` resolves to the app instead. That silently launches a whole
// Obsidian instance per call, which never answers the command and hangs until
// the timeout below fires.
const Binary = "obsidian-cli"

// defaultTimeout bounds a single CLI invocation. The CLI talks to the Obsidian
// app over IPC; when the app is hung or quitting, the call can block forever
// without this. Override with OBSIDIAN_MCP_TIMEOUT (seconds).
const defaultTimeout = 30 * time.Second

func runTimeout() time.Duration {
	if v := os.Getenv("OBSIDIAN_MCP_TIMEOUT"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil && secs > 0 {
			return time.Duration(secs) * time.Second
		}
	}
	return defaultTimeout
}

type Args struct {
	Command string
	Vault   string
	Params  map[string]string
	Flags   []string
}

func Run(ctx context.Context, a Args) (string, error) {
	if a.Command == "" {
		return "", fmt.Errorf("command required")
	}
	cliArgs := make([]string, 0, 4+len(a.Params)+len(a.Flags))
	vault := a.Vault
	if vault == "" {
		vault = os.Getenv("OBSIDIAN_DEFAULT_VAULT")
	}
	if vault != "" {
		cliArgs = append(cliArgs, "vault="+vault)
	}
	cliArgs = append(cliArgs, a.Command)
	keys := make([]string, 0, len(a.Params))
	for k := range a.Params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		cliArgs = append(cliArgs, k+"="+a.Params[k])
	}
	cliArgs = append(cliArgs, a.Flags...)

	runCtx, cancel := context.WithTimeout(ctx, runTimeout())
	defer cancel()
	cmd := exec.CommandContext(runCtx, Binary, cliArgs...)
	// On timeout the SIGKILL only reaches the CLI process itself; if it left a
	// helper child holding the stdout/stderr pipes (the obsidian CLI is the
	// Electron binary and can spawn helpers), Run would keep waiting for the
	// pipes to drain until every descendant exits. WaitDelay caps that wait.
	cmd.WaitDelay = 5 * time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	// Propagate context cancellation distinctly. exec.CommandContext kills the
	// subprocess when ctx is canceled; if we return a regular error here, the
	// SDK sends a response for a request the client already abandoned, which
	// corrupts the stdio message stream and closes the transport. Returning
	// ctx.Err() lets the SDK skip the response.
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	if runCtx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("obsidian %s: timed out after %s — is the Obsidian app running?", a.Command, runTimeout())
	}
	if err != nil {
		msg := strings.TrimSpace(stripPreamble(stderr.String()))
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("obsidian %s: %s", a.Command, msg)
	}
	out := stdout.String()
	if msg := stdoutError(out); msg != "" {
		return "", fmt.Errorf("obsidian %s: %s", a.Command, msg)
	}
	return out, nil
}

// stdoutError classifies CLI output that reports a failure despite exit code 0.
// The CLI prints single-line `Error: ...` messages to stdout for most failures
// (missing file, unknown tag, ...). Only a single-line match counts: multi-line
// output starting with "Error:" is far more likely to be real note content
// (incident logs, troubleshooting notes) returned by read/search commands.
func stdoutError(out string) string {
	trimmed := strings.TrimSpace(out)
	if strings.HasPrefix(trimmed, "Error:") && !strings.Contains(trimmed, "\n") {
		return trimmed
	}
	return ""
}

func VaultPath(ctx context.Context, vault string) (string, error) {
	out, err := Run(ctx, Args{Command: "vault", Vault: vault, Params: map[string]string{"info": "path"}})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

var preamblePatterns = []string{
	"Loading updated app package",
	"Your Obsidian installer is out of date",
	"Please download the latest installer",
}

func stripPreamble(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		skip := false
		for _, p := range preamblePatterns {
			if strings.Contains(line, p) {
				skip = true
				break
			}
		}
		if !skip {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

// CLISafe reports whether content survives the CLI `content=` round-trip.
// The CLI decodes literal \n and \t sequences into real whitespace with a
// blind replace and has no escape for the backslash itself, so content
// containing any backslash cannot be encoded losslessly — `C:\notes`,
// LaTeX `\newcommand`, and regex snippets would come back corrupted.
// Such content must be written via WriteFileDirect instead.
func CLISafe(s string) bool {
	return !strings.Contains(s, `\`)
}

// EncodeMultiline prepares CLI-safe content (see CLISafe) for the `content=`
// parameter. Callers must check CLISafe first: for content containing
// backslashes this encoding is lossy.
func EncodeMultiline(s string) string {
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}

// SecureJoin resolves rel against root and rejects any result that escapes
// root (e.g. via `..` segments).
func SecureJoin(root, rel string) (string, error) {
	abs := filepath.Join(root, rel)
	r, err := filepath.Rel(root, abs)
	if err != nil || r == ".." || strings.HasPrefix(r, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path %q escapes the vault", rel)
	}
	return abs, nil
}

// WriteFileDirect writes content straight into the vault filesystem, for
// content the CLI cannot transport losslessly (see CLISafe). Obsidian picks
// the change up through its file watcher.
func WriteFileDirect(ctx context.Context, vault, rel, content string, overwrite bool) error {
	root, err := VaultPath(ctx, vault)
	if err != nil {
		return err
	}
	abs, err := SecureJoin(root, rel)
	if err != nil {
		return err
	}
	if !overwrite {
		if _, statErr := os.Stat(abs); statErr == nil {
			return fmt.Errorf("file already exists: %s", rel)
		}
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	return os.WriteFile(abs, []byte(content), 0o644)
}
