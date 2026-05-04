package tools

import (
	"context"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TemplatesListInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Total bool   `json:"total,omitempty" jsonschema:"return only the total count of templates"`
}

func templatesHandler(ctx context.Context, _ *mcp.CallToolRequest, in TemplatesListInput) (*mcp.CallToolResult, TextOutput, error) {
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "templates", Vault: in.Vault, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type TemplateReadInput struct {
	Vault   string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Name    string `json:"name" jsonschema:"template name to read"`
	Resolve bool   `json:"resolve,omitempty" jsonschema:"resolve template variables before returning"`
	Title   string `json:"title,omitempty" jsonschema:"title used when resolving template variables"`
}

func templateReadHandler(ctx context.Context, _ *mcp.CallToolRequest, in TemplateReadInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{
		"name": in.Name,
	}
	if in.Title != "" {
		params["title"] = in.Title
	}
	flags := []string{}
	if in.Resolve {
		flags = append(flags, "resolve")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "template:read", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type TemplateInsertInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Name  string `json:"name" jsonschema:"template name to insert into the active file"`
}

func templateInsertHandler(ctx context.Context, _ *mcp.CallToolRequest, in TemplateInsertInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{
		"name": in.Name,
	}
	out, err := exec.Run(ctx, exec.Args{Command: "template:insert", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

func RegisterTemplates(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_templates",
		Description: "List available templates. total=true returns only the count.",
	}, templatesHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_template_read",
		Description: "Read a template by name. resolve=true resolves variables; title sets the title used during resolution.",
	}, templateReadHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_template_insert",
		Description: "Insert a template by name into the active file.",
	}, templateInsertHandler)
}
