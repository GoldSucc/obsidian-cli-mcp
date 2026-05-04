package tools

import (
	"context"
	"strconv"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SyncInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	On    bool   `json:"on,omitempty" jsonschema:"resume Obsidian Sync"`
	Off   bool   `json:"off,omitempty" jsonschema:"pause Obsidian Sync"`
}

func syncHandler(ctx context.Context, _ *mcp.CallToolRequest, in SyncInput) (*mcp.CallToolResult, TextOutput, error) {
	flags := []string{}
	if in.On {
		flags = append(flags, "on")
	}
	if in.Off {
		flags = append(flags, "off")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "sync", Vault: in.Vault, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type SyncStatusInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
}

func syncStatusHandler(ctx context.Context, _ *mcp.CallToolRequest, in SyncStatusInput) (*mcp.CallToolResult, TextOutput, error) {
	out, err := exec.Run(ctx, exec.Args{Command: "sync:status", Vault: in.Vault})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type SyncDeletedInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Total bool   `json:"total,omitempty" jsonschema:"only print the total count"`
}

func syncDeletedHandler(ctx context.Context, _ *mcp.CallToolRequest, in SyncDeletedInput) (*mcp.CallToolResult, TextOutput, error) {
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "sync:deleted", Vault: in.Vault, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type SyncHistoryInput struct {
	FileTarget
	Total bool `json:"total,omitempty" jsonschema:"only print the total count"`
}

func syncHistoryHandler(ctx context.Context, _ *mcp.CallToolRequest, in SyncHistoryInput) (*mcp.CallToolResult, TextOutput, error) {
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "sync:history", Vault: in.Vault, Params: in.params(), Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type SyncReadInput struct {
	FileTarget
	Version int `json:"version" jsonschema:"sync version number to read; required"`
}

func syncReadHandler(ctx context.Context, _ *mcp.CallToolRequest, in SyncReadInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	if in.Version != 0 {
		params["version"] = strconv.Itoa(in.Version)
	}
	out, err := exec.Run(ctx, exec.Args{Command: "sync:read", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type SyncRestoreInput struct {
	FileTarget
	Version int `json:"version" jsonschema:"sync version number to restore; required"`
}

func syncRestoreHandler(ctx context.Context, _ *mcp.CallToolRequest, in SyncRestoreInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	if in.Version != 0 {
		params["version"] = strconv.Itoa(in.Version)
	}
	out, err := exec.Run(ctx, exec.Args{Command: "sync:restore", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type SyncOpenInput struct {
	FileTarget
}

func syncOpenHandler(ctx context.Context, _ *mcp.CallToolRequest, in SyncOpenInput) (*mcp.CallToolResult, TextOutput, error) {
	out, err := exec.Run(ctx, exec.Args{Command: "sync:open", Vault: in.Vault, Params: in.params()})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

func RegisterSync(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_sync",
		Description: "Pause or resume Obsidian Sync. Set on=true to resume, off=true to pause.",
	}, syncHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_sync_status",
		Description: "Show Obsidian Sync status for the current vault.",
	}, syncStatusHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_sync_deleted",
		Description: "List files deleted via Sync. Set total=true to only print the count.",
	}, syncDeletedHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_sync_history",
		Description: "Show Sync version history for a file. Set total=true to only print the count.",
	}, syncHistoryHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_sync_read",
		Description: "Read a specific Sync version of a file. Version is required.",
	}, syncReadHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_sync_restore",
		Description: "Restore a file to a specific Sync version. Version is required.",
	}, syncRestoreHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_sync_open",
		Description: "Open the Sync history UI for a file.",
	}, syncOpenHandler)
}
