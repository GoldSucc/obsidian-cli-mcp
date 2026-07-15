package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type FileTarget struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	File  string `json:"file,omitempty" jsonschema:"file name resolved like a wikilink (no path/extension)"`
	Path  string `json:"path,omitempty" jsonschema:"exact path from vault root, e.g. 'folder/note.md'"`
}

func (t FileTarget) params() map[string]string {
	p := map[string]string{}
	if t.File != "" {
		p["file"] = t.File
	}
	if t.Path != "" {
		p["path"] = t.Path
	}
	return p
}

// targetPath returns the exact vault-relative path for a target. A path is
// used as-is; a file (wikilink-style) is resolved through the CLI's `file`
// command, whose first output line is `path\t<path>`. With neither set the
// CLI resolves the active file.
func targetPath(ctx context.Context, vault string, t FileTarget) (string, error) {
	if t.Path != "" {
		return t.Path, nil
	}
	params := map[string]string{}
	if t.File != "" {
		params["file"] = t.File
	}
	out, err := exec.Run(ctx, exec.Args{Command: "file", Vault: vault, Params: params})
	if err != nil {
		return "", err
	}
	for line := range strings.SplitSeq(out, "\n") {
		if rest, ok := strings.CutPrefix(line, "path\t"); ok {
			return strings.TrimSpace(rest), nil
		}
	}
	return "", fmt.Errorf("could not resolve path for target %+v", t)
}

// resolvePath is targetPath for tools that must not silently operate on the
// active file (replace, edit, insert): an explicit target is required.
func resolvePath(ctx context.Context, vault string, t FileTarget) (string, error) {
	if t.File == "" && t.Path == "" {
		return "", fmt.Errorf("file or path is required")
	}
	return targetPath(ctx, vault, t)
}

type ReadInput struct {
	FileTarget
	FrontmatterOnly bool `json:"frontmatter_only,omitempty" jsonschema:"return only the YAML frontmatter block"`
	BodyOnly        bool `json:"body_only,omitempty" jsonschema:"return only the body, without frontmatter"`
}

type TextOutput struct {
	Content string `json:"content"`
}

func readHandler(ctx context.Context, _ *mcp.CallToolRequest, in ReadInput) (*mcp.CallToolResult, TextOutput, error) {
	out, err := exec.Run(ctx, exec.Args{Command: "read", Vault: in.Vault, Params: in.params()})
	if err != nil {
		return nil, TextOutput{}, err
	}
	out = sliceNote(out, in.FrontmatterOnly, in.BodyOnly)
	return nil, TextOutput{Content: out}, nil
}

type ReadManyInput struct {
	Vault           string   `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Paths           []string `json:"paths" jsonschema:"vault-relative paths to read, in order"`
	FrontmatterOnly bool     `json:"frontmatter_only,omitempty" jsonschema:"return only each note's YAML frontmatter block"`
	BodyOnly        bool     `json:"body_only,omitempty" jsonschema:"return only each note's body, without frontmatter"`
}

func readManyHandler(ctx context.Context, _ *mcp.CallToolRequest, in ReadManyInput) (*mcp.CallToolResult, TextOutput, error) {
	if len(in.Paths) == 0 {
		return nil, TextOutput{}, fmt.Errorf("paths is required")
	}
	root, err := exec.VaultPath(ctx, in.Vault)
	if err != nil {
		return nil, TextOutput{}, err
	}
	var b strings.Builder
	for i, rel := range in.Paths {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "==> %s <==\n", rel)
		abs, jErr := exec.SecureJoin(root, rel)
		if jErr != nil {
			fmt.Fprintf(&b, "[error: %v]\n", jErr)
			continue
		}
		data, rErr := os.ReadFile(abs)
		if rErr != nil {
			fmt.Fprintf(&b, "[error: %v]\n", rErr)
			continue
		}
		content := sliceNote(string(data), in.FrontmatterOnly, in.BodyOnly)
		b.WriteString(strings.TrimRight(content, "\n"))
		b.WriteString("\n")
	}
	return nil, TextOutput{Content: b.String()}, nil
}

type CreateInput struct {
	Vault    string `json:"vault,omitempty"`
	Name     string `json:"name,omitempty" jsonschema:"file name (without extension)"`
	Path     string `json:"path,omitempty" jsonschema:"explicit path from vault root"`
	Content  string `json:"content,omitempty" jsonschema:"initial file content; real newlines are accepted"`
	Template string `json:"template,omitempty" jsonschema:"template name to use"`
	Open      bool  `json:"open,omitempty" jsonschema:"open file after creation"`
	NewTab    bool  `json:"newtab,omitempty" jsonschema:"open in new tab (implies open)"`
}

func createHandler(ctx context.Context, _ *mcp.CallToolRequest, in CreateInput) (*mcp.CallToolResult, TextOutput, error) {
	if in.Content != "" && !exec.CLISafe(in.Content) {
		return createDirect(ctx, in)
	}
	params := map[string]string{}
	if in.Name != "" {
		params["name"] = in.Name
	}
	if in.Path != "" {
		params["path"] = in.Path
	}
	if in.Content != "" {
		params["content"] = exec.EncodeMultiline(in.Content)
	}
	if in.Template != "" {
		params["template"] = in.Template
	}
	flags := []string{}
	if in.Open {
		flags = append(flags, "open")
	}
	if in.NewTab {
		flags = append(flags, "newtab")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "create", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

// createDirect handles create for content the CLI cannot transport (see
// exec.CLISafe). The note lands on disk via a direct filesystem write;
// name/template resolution still goes through the CLI when needed.
func createDirect(ctx context.Context, in CreateInput) (*mcp.CallToolResult, TextOutput, error) {
	path := in.Path
	if path == "" || in.Template != "" {
		// Let the CLI create the (empty or templated) note first so its
		// name/auto-folder/template logic applies, then locate the result.
		params := map[string]string{}
		if in.Name != "" {
			params["name"] = in.Name
		}
		if in.Path != "" {
			params["path"] = in.Path
		}
		if in.Template != "" {
			params["template"] = in.Template
		}
		out, err := exec.Run(ctx, exec.Args{Command: "create", Vault: in.Vault, Params: params})
		if err != nil {
			return nil, TextOutput{}, err
		}
		path = parseCreatedPath(out)
		if path == "" {
			return nil, TextOutput{}, fmt.Errorf("could not locate created note in CLI output: %q", strings.TrimSpace(out))
		}
		content := in.Content
		if in.Template != "" {
			existing, rErr := exec.Run(ctx, exec.Args{Command: "read", Vault: in.Vault, Params: map[string]string{"path": path}})
			if rErr == nil && strings.TrimSpace(existing) != "" {
				content = strings.TrimRight(existing, "\n") + "\n" + in.Content
			}
		}
		if err := exec.WriteFileDirect(ctx, in.Vault, path, content, true); err != nil {
			return nil, TextOutput{}, err
		}
	} else {
		if err := exec.WriteFileDirect(ctx, in.Vault, path, in.Content, false); err != nil {
			return nil, TextOutput{}, err
		}
	}
	if in.Open || in.NewTab {
		flags := []string{}
		if in.NewTab {
			flags = append(flags, "newtab")
		}
		if _, err := exec.Run(ctx, exec.Args{Command: "open", Vault: in.Vault, Params: map[string]string{"path": path}, Flags: flags}); err != nil {
			return nil, TextOutput{Content: "Created: " + path + " (open failed: " + err.Error() + ")"}, nil
		}
	}
	return nil, TextOutput{Content: "Created: " + path}, nil
}

// parseCreatedPath extracts the note path from the CLI's `Created: <path>`
// confirmation line.
func parseCreatedPath(out string) string {
	for line := range strings.SplitSeq(out, "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "Created: "); ok {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

// writeContent overwrites (or creates) a note at a resolved path. CLI-safe
// content goes through the CLI `create` command; content the CLI would
// corrupt (see exec.CLISafe) is written directly into the vault filesystem.
func writeContent(ctx context.Context, vault, path, content string, overwrite bool) (string, error) {
	if exec.CLISafe(content) {
		params := map[string]string{
			"path":    path,
			"content": exec.EncodeMultiline(content),
		}
		flags := []string{}
		if overwrite {
			flags = append(flags, "overwrite")
		}
		return exec.Run(ctx, exec.Args{Command: "create", Vault: vault, Params: params, Flags: flags})
	}
	if err := exec.WriteFileDirect(ctx, vault, path, content, overwrite); err != nil {
		return "", err
	}
	return "Wrote: " + path, nil
}

type ReplaceInput struct {
	FileTarget
	Content string `json:"content" jsonschema:"full replacement content for the entire note; real newlines accepted"`
}

func replaceHandler(ctx context.Context, _ *mcp.CallToolRequest, in ReplaceInput) (*mcp.CallToolResult, TextOutput, error) {
	path, err := resolvePath(ctx, in.Vault, in.FileTarget)
	if err != nil {
		return nil, TextOutput{}, err
	}
	out, err := writeContent(ctx, in.Vault, path, in.Content, true)
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type EditInput struct {
	FileTarget
	OldString  string `json:"old_string" jsonschema:"exact text to find; must match including whitespace and be unique unless replace_all is set"`
	NewString  string `json:"new_string" jsonschema:"text to replace it with; must differ from old_string"`
	ReplaceAll bool   `json:"replace_all,omitempty" jsonschema:"replace every occurrence instead of requiring a single unique match"`
}

func editHandler(ctx context.Context, _ *mcp.CallToolRequest, in EditInput) (*mcp.CallToolResult, TextOutput, error) {
	if in.OldString == in.NewString {
		return nil, TextOutput{}, fmt.Errorf("old_string and new_string are identical")
	}
	path, err := resolvePath(ctx, in.Vault, in.FileTarget)
	if err != nil {
		return nil, TextOutput{}, err
	}
	content, err := exec.Run(ctx, exec.Args{Command: "read", Vault: in.Vault, Params: map[string]string{"path": path}})
	if err != nil {
		return nil, TextOutput{}, err
	}
	count := strings.Count(content, in.OldString)
	if count == 0 {
		return nil, TextOutput{}, fmt.Errorf("old_string not found in %s", path)
	}
	if !in.ReplaceAll && count > 1 {
		return nil, TextOutput{}, fmt.Errorf("old_string is not unique in %s (%d matches); add more surrounding context or set replace_all", path, count)
	}
	var updated string
	if in.ReplaceAll {
		updated = strings.ReplaceAll(content, in.OldString, in.NewString)
	} else {
		updated = strings.Replace(content, in.OldString, in.NewString, 1)
	}
	if _, err := writeContent(ctx, in.Vault, path, updated, true); err != nil {
		return nil, TextOutput{}, err
	}
	noun := "match"
	if count > 1 {
		noun = "matches"
	}
	verb := "Replaced 1"
	if in.ReplaceAll {
		verb = fmt.Sprintf("Replaced %d", count)
	}
	return nil, TextOutput{Content: fmt.Sprintf("%s %s in %s", verb, noun, path)}, nil
}

type AppendInput struct {
	FileTarget
	Content string `json:"content" jsonschema:"content to append; real newlines accepted"`
	Inline  bool   `json:"inline,omitempty" jsonschema:"append without leading newline"`
}

func appendHandler(ctx context.Context, _ *mcp.CallToolRequest, in AppendInput) (*mcp.CallToolResult, TextOutput, error) {
	if !exec.CLISafe(in.Content) {
		out, err := spliceDirect(ctx, in.Vault, in.FileTarget, in.Content, in.Inline, false)
		if err != nil {
			return nil, TextOutput{}, err
		}
		return nil, TextOutput{Content: out}, nil
	}
	params := in.params()
	params["content"] = exec.EncodeMultiline(in.Content)
	flags := []string{}
	if in.Inline {
		flags = append(flags, "inline")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "append", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type PrependInput struct {
	FileTarget
	Content string `json:"content" jsonschema:"text to prepend; real newlines accepted"`
	Inline  bool   `json:"inline,omitempty" jsonschema:"prepend without trailing newline"`
}

func prependHandler(ctx context.Context, _ *mcp.CallToolRequest, in PrependInput) (*mcp.CallToolResult, TextOutput, error) {
	if !exec.CLISafe(in.Content) {
		out, err := spliceDirect(ctx, in.Vault, in.FileTarget, in.Content, in.Inline, true)
		if err != nil {
			return nil, TextOutput{}, err
		}
		return nil, TextOutput{Content: out}, nil
	}
	params := in.params()
	params["content"] = exec.EncodeMultiline(in.Content)
	flags := []string{}
	if in.Inline {
		flags = append(flags, "inline")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "prepend", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

// spliceDirect implements append/prepend for content the CLI cannot transport
// (see exec.CLISafe): read the note, concatenate, write back directly.
func spliceDirect(ctx context.Context, vault string, t FileTarget, content string, inline, prepend bool) (string, error) {
	path, err := targetPath(ctx, vault, t)
	if err != nil {
		return "", err
	}
	existing, err := exec.Run(ctx, exec.Args{Command: "read", Vault: vault, Params: map[string]string{"path": path}})
	if err != nil {
		return "", err
	}
	sep := "\n"
	if inline {
		sep = ""
	}
	var updated string
	if prepend {
		updated = content + sep + existing
	} else {
		updated = existing + sep + content
	}
	return writeContent(ctx, vault, path, updated, true)
}

type MoveInput struct {
	FileTarget
	To string `json:"to" jsonschema:"destination path from vault root"`
}

func moveHandler(ctx context.Context, _ *mcp.CallToolRequest, in MoveInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	params["to"] = in.To
	args := exec.Args{Command: "move", Vault: in.Vault, Params: params}
	out, err := exec.Run(ctx, args)
	if err != nil && strings.Contains(err.Error(), "ENOENT") && strings.Contains(err.Error(), in.To) {
		// CLI's move uses raw rename(2) — no auto-mkdir. Create destination dir, retry.
		if root, vErr := exec.VaultPath(ctx, in.Vault); vErr == nil && root != "" {
			if dest, jErr := exec.SecureJoin(root, filepath.Dir(in.To)); jErr == nil {
				if mkErr := os.MkdirAll(dest, 0o755); mkErr == nil {
					out, err = exec.Run(ctx, args)
				}
			}
		}
	}
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type RenameInput struct {
	FileTarget
	Name string `json:"name" jsonschema:"new file name"`
}

func renameHandler(ctx context.Context, _ *mcp.CallToolRequest, in RenameInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	params["name"] = in.Name
	out, err := exec.Run(ctx, exec.Args{Command: "rename", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type DeleteInput struct {
	FileTarget
	Permanent bool `json:"permanent,omitempty" jsonschema:"skip the trash and delete permanently"`
}

func deleteHandler(ctx context.Context, _ *mcp.CallToolRequest, in DeleteInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	flags := []string{}
	if in.Permanent {
		flags = append(flags, "permanent")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "delete", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type FileInfoInput struct {
	FileTarget
}

func fileInfoHandler(ctx context.Context, _ *mcp.CallToolRequest, in FileInfoInput) (*mcp.CallToolResult, TextOutput, error) {
	out, err := exec.Run(ctx, exec.Args{Command: "file", Vault: in.Vault, Params: in.params()})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type FilesListInput struct {
	Vault  string `json:"vault,omitempty"`
	Folder string `json:"folder,omitempty" jsonschema:"folder path from vault root; defaults to whole vault"`
	Ext    string `json:"ext,omitempty" jsonschema:"filter by extension, e.g. 'md'"`
	Total  bool   `json:"total,omitempty" jsonschema:"only print the total count"`
}

func filesListHandler(ctx context.Context, _ *mcp.CallToolRequest, in FilesListInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.Folder != "" {
		params["folder"] = in.Folder
	}
	if in.Ext != "" {
		params["ext"] = in.Ext
	}
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "files", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type WordCountInput struct {
	FileTarget
	Words      bool `json:"words,omitempty" jsonschema:"count words only"`
	Characters bool `json:"characters,omitempty" jsonschema:"count characters only"`
}

func wordCountHandler(ctx context.Context, _ *mcp.CallToolRequest, in WordCountInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	flags := []string{}
	if in.Words {
		flags = append(flags, "words")
	}
	if in.Characters {
		flags = append(flags, "characters")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "wordcount", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

func RegisterFiles(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_read",
		Description: "Read full contents of a note. Specify file (wikilink-style) or path (exact). Omit both for active file. frontmatter_only/body_only narrow the result.",
	}, readHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_read_many",
		Description: "Read several notes in one call. Provide vault-relative paths; returns each note prefixed by '==> path <=='. frontmatter_only/body_only narrow each note.",
	}, readManyHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_create",
		Description: "Create a NEW note. Use name (auto-folder by templates) or path (explicit). Optional content, template, open/newtab flags. Fails if the note already exists — use obsidian_edit (partial change) or obsidian_replace (full rewrite) to modify existing notes.",
	}, createHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_replace",
		Description: "Replace the ENTIRE content of an existing note with new content. Target by file (wikilink-style) or path (exact). For changing only part of a note, prefer obsidian_edit.",
	}, replaceHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_edit",
		Description: "Change PART of a note via exact string replacement (Claude-Edit style). Provide old_string (must match exactly, including whitespace, and be unique unless replace_all=true) and new_string. Target by file (wikilink-style) or path (exact).",
	}, editHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_append",
		Description: "Append content to a note. inline=true skips the leading newline.",
	}, appendHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_prepend",
		Description: "Prepend content to a note. inline=true skips the trailing newline.",
	}, prependHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_move",
		Description: "Move a note to a new path inside the vault.",
	}, moveHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_rename",
		Description: "Rename a note in place. Provide the new name (without path).",
	}, renameHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_delete",
		Description: "Delete a note. Defaults to trash; permanent=true bypasses the trash.",
	}, deleteHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_file_info",
		Description: "Show metadata for a single note (path, size, dates, etc).",
	}, fileInfoHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_files_list",
		Description: "List files in the vault. Optional folder scope, extension filter, total-only flag.",
	}, filesListHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_wordcount",
		Description: "Count words and/or characters in a note. Set words or characters to narrow the result.",
	}, wordCountHandler)
}
