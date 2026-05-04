package tools

import (
	"context"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type WorkspaceInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Ids   bool   `json:"ids,omitempty" jsonschema:"include ids in the output"`
}

func workspaceHandler(ctx context.Context, _ *mcp.CallToolRequest, in WorkspaceInput) (*mcp.CallToolResult, TextOutput, error) {
	flags := []string{}
	if in.Ids {
		flags = append(flags, "ids")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "workspace", Vault: in.Vault, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type TabsInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Ids   bool   `json:"ids,omitempty" jsonschema:"include ids in the output"`
}

func tabsHandler(ctx context.Context, _ *mcp.CallToolRequest, in TabsInput) (*mcp.CallToolResult, TextOutput, error) {
	flags := []string{}
	if in.Ids {
		flags = append(flags, "ids")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "tabs", Vault: in.Vault, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type TabOpenInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Group string `json:"group,omitempty" jsonschema:"target tab group id"`
	File  string `json:"file,omitempty" jsonschema:"file path to open"`
	View  string `json:"view,omitempty" jsonschema:"view type to open"`
}

func tabOpenHandler(ctx context.Context, _ *mcp.CallToolRequest, in TabOpenInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.Group != "" {
		params["group"] = in.Group
	}
	if in.File != "" {
		params["file"] = in.File
	}
	if in.View != "" {
		params["view"] = in.View
	}
	out, err := exec.Run(ctx, exec.Args{Command: "tab:open", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type RecentsInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Total bool   `json:"total,omitempty" jsonschema:"include total count"`
}

func recentsHandler(ctx context.Context, _ *mcp.CallToolRequest, in RecentsInput) (*mcp.CallToolResult, TextOutput, error) {
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "recents", Vault: in.Vault, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

func RegisterWorkspace(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_workspace",
		Description: "Show current workspace layout. Use ids to include element ids.",
	}, workspaceHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_tabs",
		Description: "List open tabs. Use ids to include tab ids.",
	}, tabsHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_tab_open",
		Description: "Open a file or view in a tab group.",
	}, tabOpenHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_recents",
		Description: "List recently opened files. Use total to include count.",
	}, recentsHandler)
}
