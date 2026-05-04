package tools

import (
	"context"
	"strconv"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SearchInput struct {
	Vault  string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Query  string `json:"query" jsonschema:"search query text"`
	Path   string `json:"path,omitempty" jsonschema:"restrict search to a folder under the vault root"`
	Limit  int    `json:"limit,omitempty" jsonschema:"maximum number of results to return"`
	Total  bool   `json:"total,omitempty" jsonschema:"include only the total count of matches"`
	Case   bool   `json:"case,omitempty" jsonschema:"case-sensitive matching"`
	Format string `json:"format,omitempty" jsonschema:"output format: text or json"`
}

func searchHandler(ctx context.Context, _ *mcp.CallToolRequest, in SearchInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.Query != "" {
		params["query"] = in.Query
	}
	if in.Path != "" {
		params["path"] = in.Path
	}
	if in.Limit != 0 {
		params["limit"] = strconv.Itoa(in.Limit)
	}
	if in.Format != "" {
		params["format"] = in.Format
	}
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	if in.Case {
		flags = append(flags, "case")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "search", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type SearchContextInput struct {
	Vault  string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Query  string `json:"query" jsonschema:"search query text"`
	Path   string `json:"path,omitempty" jsonschema:"restrict search to a folder under the vault root"`
	Limit  int    `json:"limit,omitempty" jsonschema:"maximum number of results to return"`
	Case   bool   `json:"case,omitempty" jsonschema:"case-sensitive matching"`
	Format string `json:"format,omitempty" jsonschema:"output format: text or json"`
}

func searchContextHandler(ctx context.Context, _ *mcp.CallToolRequest, in SearchContextInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.Query != "" {
		params["query"] = in.Query
	}
	if in.Path != "" {
		params["path"] = in.Path
	}
	if in.Limit != 0 {
		params["limit"] = strconv.Itoa(in.Limit)
	}
	if in.Format != "" {
		params["format"] = in.Format
	}
	flags := []string{}
	if in.Case {
		flags = append(flags, "case")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "search:context", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type SearchOpenInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Query string `json:"query,omitempty" jsonschema:"search query text to prefill the panel"`
}

func searchOpenHandler(ctx context.Context, _ *mcp.CallToolRequest, in SearchOpenInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.Query != "" {
		params["query"] = in.Query
	}
	out, err := exec.Run(ctx, exec.Args{Command: "search:open", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

func RegisterSearch(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_search",
		Description: "Search vault for a query. Optional path scope, limit, case-sensitive, total-only, or json/text format.",
	}, searchHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_search_context",
		Description: "Search vault and return matching lines with surrounding context. Optional path, limit, case, format.",
	}, searchContextHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_search_open",
		Description: "Open the Obsidian search panel, optionally prefilled with a query.",
	}, searchOpenHandler)
}
