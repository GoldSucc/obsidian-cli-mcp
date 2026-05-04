package tools

import (
	"context"
	"strconv"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type HistoryInput struct {
	FileTarget
}

func historyHandler(ctx context.Context, _ *mcp.CallToolRequest, in HistoryInput) (*mcp.CallToolResult, TextOutput, error) {
	out, err := exec.Run(ctx, exec.Args{Command: "history", Vault: in.Vault, Params: in.params()})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type HistoryListInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
}

func historyListHandler(ctx context.Context, _ *mcp.CallToolRequest, in HistoryListInput) (*mcp.CallToolResult, TextOutput, error) {
	out, err := exec.Run(ctx, exec.Args{Command: "history:list", Vault: in.Vault})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type HistoryReadInput struct {
	FileTarget
	Version int `json:"version,omitempty" jsonschema:"version number to read; defaults to 1"`
}

func historyReadHandler(ctx context.Context, _ *mcp.CallToolRequest, in HistoryReadInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	if in.Version != 0 {
		params["version"] = strconv.Itoa(in.Version)
	}
	out, err := exec.Run(ctx, exec.Args{Command: "history:read", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type HistoryRestoreInput struct {
	FileTarget
	Version int `json:"version" jsonschema:"version number to restore; required"`
}

func historyRestoreHandler(ctx context.Context, _ *mcp.CallToolRequest, in HistoryRestoreInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	if in.Version != 0 {
		params["version"] = strconv.Itoa(in.Version)
	}
	out, err := exec.Run(ctx, exec.Args{Command: "history:restore", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type HistoryOpenInput struct {
	FileTarget
}

func historyOpenHandler(ctx context.Context, _ *mcp.CallToolRequest, in HistoryOpenInput) (*mcp.CallToolResult, TextOutput, error) {
	out, err := exec.Run(ctx, exec.Args{Command: "history:open", Vault: in.Vault, Params: in.params()})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type DiffInput struct {
	FileTarget
	From   int    `json:"from,omitempty" jsonschema:"version number to diff from"`
	To     int    `json:"to,omitempty" jsonschema:"version number to diff to"`
	Filter string `json:"filter,omitempty" jsonschema:"history source filter: local or sync"`
}

func diffHandler(ctx context.Context, _ *mcp.CallToolRequest, in DiffInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	if in.From != 0 {
		params["from"] = strconv.Itoa(in.From)
	}
	if in.To != 0 {
		params["to"] = strconv.Itoa(in.To)
	}
	if in.Filter != "" {
		params["filter"] = in.Filter
	}
	out, err := exec.Run(ctx, exec.Args{Command: "diff", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

func RegisterHistory(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_history",
		Description: "Show the version history for a file. Specify file (wikilink-style) or path.",
	}, historyHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_history_list",
		Description: "List all files that have local version history.",
	}, historyListHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_history_read",
		Description: "Read a specific version from a file's history. Defaults to version 1.",
	}, historyReadHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_history_restore",
		Description: "Restore a file to a specific version from its history. Version is required.",
	}, historyRestoreHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_history_open",
		Description: "Open Obsidian's file recovery UI for the given file.",
	}, historyOpenHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_diff",
		Description: "List versions or diff two versions of a file. Optional from/to versions and filter (local|sync).",
	}, diffHandler)
}
