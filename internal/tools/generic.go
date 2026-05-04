package tools

import (
	"context"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type RunInput struct {
	Command string            `json:"command" jsonschema:"obsidian CLI command, e.g. 'read', 'daily:append', 'plugin:reload'"`
	Vault   string            `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT env or most recently focused vault"`
	Params  map[string]string `json:"params,omitempty" jsonschema:"CLI parameters as name/value pairs, e.g. {\"file\":\"My Note\",\"content\":\"Hi\"}"`
	Flags   []string          `json:"flags,omitempty" jsonschema:"boolean flags without value, e.g. [\"silent\",\"overwrite\"]"`
}

type RunOutput struct {
	Stdout string `json:"stdout"`
}

func runHandler(ctx context.Context, _ *mcp.CallToolRequest, in RunInput) (*mcp.CallToolResult, RunOutput, error) {
	out, err := exec.Run(ctx, exec.Args{
		Command: in.Command,
		Vault:   in.Vault,
		Params:  in.Params,
		Flags:   in.Flags,
	})
	if err != nil {
		return nil, RunOutput{}, err
	}
	return nil, RunOutput{Stdout: out}, nil
}

func RegisterGeneric(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_run",
		Description: "Escape hatch: run any Obsidian CLI command not covered by typed tools. Use for plugin/theme/dev commands or new commands not yet wrapped. See `obsidian help` for the full command list.",
	}, runHandler)
}
