package exec

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

const Binary = "obsidian"

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

	cmd := exec.CommandContext(ctx, Binary, cliArgs...)
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
	if err != nil {
		msg := strings.TrimSpace(stripPreamble(stderr.String()))
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("obsidian %s: %s", a.Command, msg)
	}
	out := stdout.String()
	// CLI reports many failures via stdout with exit code 0, prefixed `Error:`.
	if trimmed := strings.TrimSpace(out); strings.HasPrefix(trimmed, "Error:") {
		return "", fmt.Errorf("obsidian %s: %s", a.Command, trimmed)
	}
	return out, nil
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

func EncodeMultiline(s string) string {
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}
