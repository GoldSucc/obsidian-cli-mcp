package tools

import (
	"context"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RandomOpenInput struct {
	Vault  string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Folder string `json:"folder,omitempty" jsonschema:"restrict random pick to this folder (path from vault root)"`
	NewTab bool   `json:"newtab,omitempty" jsonschema:"open in a new tab"`
}

func randomHandler(ctx context.Context, _ *mcp.CallToolRequest, in RandomOpenInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.Folder != "" {
		params["folder"] = in.Folder
	}
	flags := []string{}
	if in.NewTab {
		flags = append(flags, "newtab")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "random", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type RandomReadInput struct {
	Vault  string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Folder string `json:"folder,omitempty" jsonschema:"restrict random pick to this folder (path from vault root)"`
}

func randomReadHandler(ctx context.Context, _ *mcp.CallToolRequest, in RandomReadInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.Folder != "" {
		params["folder"] = in.Folder
	}
	out, err := exec.Run(ctx, exec.Args{Command: "random:read", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

func RegisterRandom(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_random",
		Description: "Open a random note. Optional folder restricts the pick; newtab opens in a new tab.",
	}, randomHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_random_read",
		Description: "Return the contents of a random note. Optional folder restricts the pick.",
	}, randomReadHandler)
}
