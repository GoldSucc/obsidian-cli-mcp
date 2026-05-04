package tools

import (
	"context"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type BookmarkInput struct {
	Vault   string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	File    string `json:"file,omitempty" jsonschema:"path of a file to bookmark"`
	Subpath string `json:"subpath,omitempty" jsonschema:"heading or block subpath within the file"`
	Folder  string `json:"folder,omitempty" jsonschema:"folder path to bookmark"`
	Search  string `json:"search,omitempty" jsonschema:"saved search query to bookmark"`
	URL     string `json:"url,omitempty" jsonschema:"external URL to bookmark"`
	Title   string `json:"title,omitempty" jsonschema:"display title for the bookmark"`
}

func bookmarkHandler(ctx context.Context, _ *mcp.CallToolRequest, in BookmarkInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.File != "" {
		params["file"] = in.File
	}
	if in.Subpath != "" {
		params["subpath"] = in.Subpath
	}
	if in.Folder != "" {
		params["folder"] = in.Folder
	}
	if in.Search != "" {
		params["search"] = in.Search
	}
	if in.URL != "" {
		params["url"] = in.URL
	}
	if in.Title != "" {
		params["title"] = in.Title
	}
	out, err := exec.Run(ctx, exec.Args{Command: "bookmark", Vault: in.Vault, Params: params})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type BookmarksInput struct {
	Vault   string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Total   bool   `json:"total,omitempty" jsonschema:"include only the total count"`
	Verbose bool   `json:"verbose,omitempty" jsonschema:"include extra bookmark details"`
	Format  string `json:"format,omitempty" jsonschema:"output format: json, tsv, or csv"`
}

func bookmarksHandler(ctx context.Context, _ *mcp.CallToolRequest, in BookmarksInput) (*mcp.CallToolResult, TextOutput, error) {
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
	out, err := exec.Run(ctx, exec.Args{Command: "bookmarks", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

func RegisterBookmarks(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_bookmark",
		Description: "Add a bookmark for a file (optional subpath), folder, saved search, or URL. Optional title.",
	}, bookmarkHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_bookmarks",
		Description: "List bookmarks. Optional total, verbose, or format json/tsv/csv.",
	}, bookmarksHandler)
}
