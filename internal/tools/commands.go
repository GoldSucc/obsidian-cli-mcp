package tools

import (
	"context"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type CommandInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	ID    string `json:"id" jsonschema:"command id to execute, e.g. 'editor:toggle-bold'"`
}

func commandHandler(ctx context.Context, _ *mcp.CallToolRequest, in CommandInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.ID != "" {
		params["id"] = in.ID
	}
	out, err := exec.Run(ctx, exec.Args{Command: "command", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type CommandsInput struct {
	Vault  string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Filter string `json:"filter,omitempty" jsonschema:"prefix to filter command ids"`
}

func commandsHandler(ctx context.Context, _ *mcp.CallToolRequest, in CommandsInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.Filter != "" {
		params["filter"] = in.Filter
	}
	out, err := exec.Run(ctx, exec.Args{Command: "commands", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type HotkeyInput struct {
	Vault   string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	ID      string `json:"id" jsonschema:"command id to look up the hotkey for"`
	Verbose bool   `json:"verbose,omitempty" jsonschema:"include extra detail in the output"`
}

func hotkeyHandler(ctx context.Context, _ *mcp.CallToolRequest, in HotkeyInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.ID != "" {
		params["id"] = in.ID
	}
	flags := []string{}
	if in.Verbose {
		flags = append(flags, "verbose")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "hotkey", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type HotkeysInput struct {
	Vault   string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Total   bool   `json:"total,omitempty" jsonschema:"only print the total count"`
	Verbose bool   `json:"verbose,omitempty" jsonschema:"include extra detail in the output"`
	All     bool   `json:"all,omitempty" jsonschema:"include all commands, not just those with hotkeys"`
	Format  string `json:"format,omitempty" jsonschema:"output format: json, tsv, or csv"`
}

func hotkeysHandler(ctx context.Context, _ *mcp.CallToolRequest, in HotkeysInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.Format != "" {
		params["format"] = in.Format
	}
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	if in.Verbose {
		flags = append(flags, "verbose")
	}
	if in.All {
		flags = append(flags, "all")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "hotkeys", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

func RegisterCommands(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_command",
		Description: "Execute an Obsidian command by id (e.g. 'editor:toggle-bold').",
	}, commandHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_commands",
		Description: "List available Obsidian commands. Optional filter is a prefix on the command id.",
	}, commandsHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_hotkey",
		Description: "Show the hotkey bound to a command id. verbose=true adds extra detail.",
	}, hotkeyHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_hotkeys",
		Description: "List hotkeys. total prints count only; verbose adds detail; all includes commands without hotkeys; format selects json/tsv/csv.",
	}, hotkeysHandler)
}
