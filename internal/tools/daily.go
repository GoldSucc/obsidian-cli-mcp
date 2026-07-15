package tools

import (
	"context"
	"strings"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type DailyOpenInput struct {
	Vault    string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	PaneType string `json:"paneType,omitempty" jsonschema:"pane to open in: tab, split, or window"`
}

func dailyHandler(ctx context.Context, _ *mcp.CallToolRequest, in DailyOpenInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.PaneType != "" {
		params["paneType"] = in.PaneType
	}
	out, err := exec.Run(ctx, exec.Args{Command: "daily", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type DailyVaultOnlyInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
}

func dailyReadHandler(ctx context.Context, _ *mcp.CallToolRequest, in DailyVaultOnlyInput) (*mcp.CallToolResult, TextOutput, error) {
	out, err := exec.Run(ctx, exec.Args{Command: "daily:read", Vault: in.Vault})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

func dailyPathHandler(ctx context.Context, _ *mcp.CallToolRequest, in DailyVaultOnlyInput) (*mcp.CallToolResult, TextOutput, error) {
	out, err := exec.Run(ctx, exec.Args{Command: "daily:path", Vault: in.Vault})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type DailyWriteInput struct {
	Vault    string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Content  string `json:"content" jsonschema:"text to write into the daily note; real newlines accepted"`
	Inline   bool   `json:"inline,omitempty" jsonschema:"omit the leading/trailing newline"`
	Open     bool   `json:"open,omitempty" jsonschema:"open the daily note after writing"`
	PaneType string `json:"paneType,omitempty" jsonschema:"pane to open in: tab, split, or window"`
}

// dailySpliceDirect implements daily append/prepend for content the CLI
// cannot transport (see exec.CLISafe): resolve today's path, read, concatenate,
// write back directly. A missing daily note is treated as empty.
func dailySpliceDirect(ctx context.Context, in DailyWriteInput, prepend bool) (*mcp.CallToolResult, TextOutput, error) {
	pathOut, err := exec.Run(ctx, exec.Args{Command: "daily:path", Vault: in.Vault})
	if err != nil {
		return nil, TextOutput{}, err
	}
	path := strings.TrimSpace(pathOut)
	existing, err := exec.Run(ctx, exec.Args{Command: "read", Vault: in.Vault, Params: map[string]string{"path": path}})
	if err != nil {
		existing = ""
	}
	sep := "\n"
	if in.Inline || existing == "" {
		sep = ""
	}
	var updated string
	if prepend {
		updated = in.Content + sep + existing
	} else {
		updated = existing + sep + in.Content
	}
	out, err := writeContent(ctx, in.Vault, path, updated, true)
	if err != nil {
		return nil, TextOutput{}, err
	}
	if in.Open {
		if _, oErr := exec.Run(ctx, exec.Args{Command: "open", Vault: in.Vault, Params: map[string]string{"path": path}}); oErr != nil {
			out += " (open failed: " + oErr.Error() + ")"
		}
	}
	return nil, TextOutput{Content: out}, nil
}

func dailyAppendHandler(ctx context.Context, _ *mcp.CallToolRequest, in DailyWriteInput) (*mcp.CallToolResult, TextOutput, error) {
	if !exec.CLISafe(in.Content) {
		return dailySpliceDirect(ctx, in, false)
	}
	params := map[string]string{
		"content": exec.EncodeMultiline(in.Content),
	}
	if in.PaneType != "" {
		params["paneType"] = in.PaneType
	}
	flags := []string{}
	if in.Inline {
		flags = append(flags, "inline")
	}
	if in.Open {
		flags = append(flags, "open")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "daily:append", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

func dailyPrependHandler(ctx context.Context, _ *mcp.CallToolRequest, in DailyWriteInput) (*mcp.CallToolResult, TextOutput, error) {
	if !exec.CLISafe(in.Content) {
		return dailySpliceDirect(ctx, in, true)
	}
	params := map[string]string{
		"content": exec.EncodeMultiline(in.Content),
	}
	if in.PaneType != "" {
		params["paneType"] = in.PaneType
	}
	flags := []string{}
	if in.Inline {
		flags = append(flags, "inline")
	}
	if in.Open {
		flags = append(flags, "open")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "daily:prepend", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

func RegisterDaily(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_daily",
		Description: "Open today's daily note. Optional paneType selects tab, split, or window.",
	}, dailyHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_daily_read",
		Description: "Return the contents of today's daily note.",
	}, dailyReadHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_daily_append",
		Description: "Append content to today's daily note. inline=true skips the leading newline; open=true opens the note.",
	}, dailyAppendHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_daily_prepend",
		Description: "Prepend content to today's daily note. inline=true skips the trailing newline; open=true opens the note.",
	}, dailyPrependHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_daily_path",
		Description: "Return the path of today's daily note.",
	}, dailyPathHandler)
}
