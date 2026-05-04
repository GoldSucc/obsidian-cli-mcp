package tools

import (
	"context"

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

type ReadInput struct {
	FileTarget
}

type TextOutput struct {
	Content string `json:"content"`
}

func readHandler(ctx context.Context, _ *mcp.CallToolRequest, in ReadInput) (*mcp.CallToolResult, TextOutput, error) {
	out, err := exec.Run(ctx, exec.Args{Command: "read", Vault: in.Vault, Params: in.params()})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type CreateInput struct {
	Vault    string `json:"vault,omitempty"`
	Name     string `json:"name,omitempty" jsonschema:"file name (without extension)"`
	Path     string `json:"path,omitempty" jsonschema:"explicit path from vault root"`
	Content  string `json:"content,omitempty" jsonschema:"initial file content; real newlines are accepted"`
	Template string `json:"template,omitempty" jsonschema:"template name to use"`
	Overwrite bool  `json:"overwrite,omitempty" jsonschema:"overwrite if file exists"`
	Open      bool  `json:"open,omitempty" jsonschema:"open file after creation"`
	NewTab    bool  `json:"newtab,omitempty" jsonschema:"open in new tab (implies open)"`
}

func createHandler(ctx context.Context, _ *mcp.CallToolRequest, in CreateInput) (*mcp.CallToolResult, TextOutput, error) {
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
	if in.Overwrite {
		flags = append(flags, "overwrite")
	}
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

type AppendInput struct {
	FileTarget
	Content string `json:"content" jsonschema:"content to append; real newlines accepted"`
	Inline  bool   `json:"inline,omitempty" jsonschema:"append without leading newline"`
}

func appendHandler(ctx context.Context, _ *mcp.CallToolRequest, in AppendInput) (*mcp.CallToolResult, TextOutput, error) {
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

type MoveInput struct {
	FileTarget
	To string `json:"to" jsonschema:"destination path from vault root"`
}

func moveHandler(ctx context.Context, _ *mcp.CallToolRequest, in MoveInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	params["to"] = in.To
	out, err := exec.Run(ctx, exec.Args{Command: "move", Vault: in.Vault, Params: params})
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
		Description: "Read full contents of a note. Specify file (wikilink-style) or path (exact). Omit both for active file.",
	}, readHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_create",
		Description: "Create a new note. Use name (auto-folder by templates) or path (explicit). Optional content, template, overwrite/open/newtab flags.",
	}, createHandler)
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
