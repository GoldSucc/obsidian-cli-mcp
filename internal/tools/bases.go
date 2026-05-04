package tools

import (
	"context"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type BasesListInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
}

func basesListHandler(ctx context.Context, _ *mcp.CallToolRequest, in BasesListInput) (*mcp.CallToolResult, TextOutput, error) {
	out, err := exec.Run(ctx, exec.Args{Command: "bases", Vault: in.Vault})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type BaseViewsInput struct {
	FileTarget
}

func baseViewsHandler(ctx context.Context, _ *mcp.CallToolRequest, in BaseViewsInput) (*mcp.CallToolResult, TextOutput, error) {
	out, err := exec.Run(ctx, exec.Args{Command: "base:views", Vault: in.Vault, Params: in.params()})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type BaseQueryInput struct {
	FileTarget
	View   string `json:"view,omitempty" jsonschema:"view name within the base"`
	Format string `json:"format,omitempty" jsonschema:"output format: json, csv, tsv, md, or paths"`
}

func baseQueryHandler(ctx context.Context, _ *mcp.CallToolRequest, in BaseQueryInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	if in.View != "" {
		params["view"] = in.View
	}
	if in.Format != "" {
		params["format"] = in.Format
	}
	out, err := exec.Run(ctx, exec.Args{Command: "base:query", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type BaseCreateInput struct {
	FileTarget
	View    string `json:"view,omitempty" jsonschema:"view name within the base"`
	Name    string `json:"name,omitempty" jsonschema:"new item name"`
	Content string `json:"content,omitempty" jsonschema:"initial content for the new item; real newlines accepted"`
	Open    bool   `json:"open,omitempty" jsonschema:"open the new item after creation"`
	NewTab  bool   `json:"newtab,omitempty" jsonschema:"open in a new tab (implies open)"`
}

func baseCreateHandler(ctx context.Context, _ *mcp.CallToolRequest, in BaseCreateInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	if in.View != "" {
		params["view"] = in.View
	}
	if in.Name != "" {
		params["name"] = in.Name
	}
	if in.Content != "" {
		params["content"] = exec.EncodeMultiline(in.Content)
	}
	flags := []string{}
	if in.Open {
		flags = append(flags, "open")
	}
	if in.NewTab {
		flags = append(flags, "newtab")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "base:create", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

func RegisterBases(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_bases",
		Description: "List all base files in the vault.",
	}, basesListHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_base_views",
		Description: "List the views defined in a base file. Specify file (wikilink-style) or path.",
	}, baseViewsHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_base_query",
		Description: "Query a base view. Specify file or path, view name, and output format (json, csv, tsv, md, paths).",
	}, baseQueryHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_base_create",
		Description: "Create a new item in a base view. Specify file or path, view, name, optional content, open/newtab flags.",
	}, baseCreateHandler)
}
