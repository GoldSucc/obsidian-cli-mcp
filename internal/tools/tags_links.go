package tools

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TagInput struct {
	Vault   string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Name    string `json:"name" jsonschema:"tag name (without leading #)"`
	Total   bool   `json:"total,omitempty" jsonschema:"only print the total count"`
	Verbose bool   `json:"verbose,omitempty" jsonschema:"include extra detail in the output"`
}

func tagHandler(ctx context.Context, _ *mcp.CallToolRequest, in TagInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{"name": in.Name}
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	if in.Verbose {
		flags = append(flags, "verbose")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "tag", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type TagsListInput struct {
	FileTarget
	Total  bool   `json:"total,omitempty" jsonschema:"only print the total count"`
	Counts bool   `json:"counts,omitempty" jsonschema:"include occurrence counts per tag"`
	Active bool   `json:"active,omitempty" jsonschema:"limit scope to the active file"`
	Sort   string `json:"sort,omitempty" jsonschema:"sort order, e.g. 'count'"`
	Format string `json:"format,omitempty" jsonschema:"output format: json, tsv, or csv"`
}

func tagsListHandler(ctx context.Context, _ *mcp.CallToolRequest, in TagsListInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	if in.Sort != "" {
		params["sort"] = in.Sort
	}
	if in.Format != "" {
		params["format"] = in.Format
	}
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	if in.Counts {
		flags = append(flags, "counts")
	}
	if in.Active {
		flags = append(flags, "active")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "tags", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type AliasesInput struct {
	FileTarget
	Total   bool `json:"total,omitempty" jsonschema:"only print the total count"`
	Verbose bool `json:"verbose,omitempty" jsonschema:"include extra detail in the output"`
	Active  bool `json:"active,omitempty" jsonschema:"limit scope to the active file"`
}

func aliasesHandler(ctx context.Context, _ *mcp.CallToolRequest, in AliasesInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	if in.Verbose {
		flags = append(flags, "verbose")
	}
	if in.Active {
		flags = append(flags, "active")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "aliases", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type LinksInput struct {
	FileTarget
	Total bool `json:"total,omitempty" jsonschema:"only print the total count"`
}

func linksHandler(ctx context.Context, _ *mcp.CallToolRequest, in LinksInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "links", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type BacklinksInput struct {
	FileTarget
	Counts bool   `json:"counts,omitempty" jsonschema:"include occurrence counts per source"`
	Total  bool   `json:"total,omitempty" jsonschema:"only print the total count"`
	Format string `json:"format,omitempty" jsonschema:"output format: json, tsv, or csv"`
}

func backlinksHandler(ctx context.Context, _ *mcp.CallToolRequest, in BacklinksInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	if in.Format != "" {
		params["format"] = in.Format
	}
	flags := []string{}
	if in.Counts {
		flags = append(flags, "counts")
	}
	if in.Total {
		flags = append(flags, "total")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "backlinks", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type UnresolvedInput struct {
	Vault   string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Total   bool   `json:"total,omitempty" jsonschema:"only print the total count"`
	Counts  bool   `json:"counts,omitempty" jsonschema:"include occurrence counts per unresolved target"`
	Verbose bool   `json:"verbose,omitempty" jsonschema:"include extra detail in the output"`
	Format  string `json:"format,omitempty" jsonschema:"output format: json, tsv, or csv"`
}

func unresolvedHandler(ctx context.Context, _ *mcp.CallToolRequest, in UnresolvedInput) (*mcp.CallToolResult, TextOutput, error) {
	params := map[string]string{}
	if in.Format != "" {
		params["format"] = in.Format
	}
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	if in.Counts {
		flags = append(flags, "counts")
	}
	if in.Verbose {
		flags = append(flags, "verbose")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "unresolved", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type OrphansInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Total bool   `json:"total,omitempty" jsonschema:"only print the total count"`
	All   bool   `json:"all,omitempty" jsonschema:"include non-markdown files"`
}

func orphansHandler(ctx context.Context, _ *mcp.CallToolRequest, in OrphansInput) (*mcp.CallToolResult, TextOutput, error) {
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	if in.All {
		flags = append(flags, "all")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "orphans", Vault: in.Vault, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type DeadendsInput struct {
	Vault string `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Total bool   `json:"total,omitempty" jsonschema:"only print the total count"`
	All   bool   `json:"all,omitempty" jsonschema:"include non-markdown files"`
}

func deadendsHandler(ctx context.Context, _ *mcp.CallToolRequest, in DeadendsInput) (*mcp.CallToolResult, TextOutput, error) {
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	if in.All {
		flags = append(flags, "all")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "deadends", Vault: in.Vault, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type OutlineInput struct {
	FileTarget
	Total  bool   `json:"total,omitempty" jsonschema:"only print the total count"`
	Format string `json:"format,omitempty" jsonschema:"output format: tree, md, or json"`
}

func outlineHandler(ctx context.Context, _ *mcp.CallToolRequest, in OutlineInput) (*mcp.CallToolResult, TextOutput, error) {
	params := in.params()
	if in.Format != "" {
		params["format"] = in.Format
	}
	flags := []string{}
	if in.Total {
		flags = append(flags, "total")
	}
	out, err := exec.Run(ctx, exec.Args{Command: "outline", Vault: in.Vault, Params: params, Flags: flags})
	if err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: out}, nil
}

type TagQueryInput struct {
	Vault      string   `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	All        []string `json:"all,omitempty" jsonschema:"tags a note must ALL carry (AND); no leading #"`
	Any        []string `json:"any,omitempty" jsonschema:"tags of which a note must carry at least one (OR); no leading #"`
	None       []string `json:"none,omitempty" jsonschema:"tags a note must NOT carry; no leading #"`
	PathPrefix string   `json:"path_prefix,omitempty" jsonschema:"restrict results to paths under this folder"`
}

// tagFiles returns the set of note paths carrying a tag. An unknown tag is an
// empty set, not an error, so set algebra over it stays meaningful.
func tagFiles(ctx context.Context, vault, name string) (map[string]bool, error) {
	out, err := exec.Run(ctx, exec.Args{Command: "tag", Vault: vault, Params: map[string]string{"name": name}, Flags: []string{"verbose"}})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return map[string]bool{}, nil
		}
		return nil, err
	}
	files := map[string]bool{}
	for line := range strings.SplitSeq(out, "\n") {
		line = strings.TrimSpace(line)
		// verbose output opens with a `#tag<TAB>count` header line
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		files[line] = true
	}
	return files, nil
}

func tagQueryHandler(ctx context.Context, _ *mcp.CallToolRequest, in TagQueryInput) (*mcp.CallToolResult, TextOutput, error) {
	if len(in.All) == 0 && len(in.Any) == 0 {
		return nil, TextOutput{}, fmt.Errorf("at least one tag in all or any is required")
	}
	var result map[string]bool
	for _, tag := range in.All {
		files, err := tagFiles(ctx, in.Vault, tag)
		if err != nil {
			return nil, TextOutput{}, err
		}
		if result == nil {
			result = files
			continue
		}
		for f := range result {
			if !files[f] {
				delete(result, f)
			}
		}
	}
	if len(in.Any) > 0 {
		union := map[string]bool{}
		for _, tag := range in.Any {
			files, err := tagFiles(ctx, in.Vault, tag)
			if err != nil {
				return nil, TextOutput{}, err
			}
			for f := range files {
				union[f] = true
			}
		}
		if result == nil {
			result = union
		} else {
			for f := range result {
				if !union[f] {
					delete(result, f)
				}
			}
		}
	}
	for _, tag := range in.None {
		files, err := tagFiles(ctx, in.Vault, tag)
		if err != nil {
			return nil, TextOutput{}, err
		}
		for f := range files {
			delete(result, f)
		}
	}
	paths := make([]string, 0, len(result))
	for f := range result {
		if in.PathPrefix != "" && !strings.HasPrefix(f, strings.TrimSuffix(in.PathPrefix, "/")+"/") {
			continue
		}
		paths = append(paths, f)
	}
	sort.Strings(paths)
	unit := "notes"
	if len(paths) == 1 {
		unit = "note"
	}
	header := fmt.Sprintf("%d %s", len(paths), unit)
	if len(paths) == 0 {
		return nil, TextOutput{Content: header}, nil
	}
	return nil, TextOutput{Content: header + "\n" + strings.Join(paths, "\n")}, nil
}

func RegisterTagsLinks(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_tag",
		Description: "Show information about a single tag. Required name (no leading #); optional total/verbose flags.",
	}, tagHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_tags",
		Description: "List tags in the vault or a specific note. Optional file/path scope, total/counts/active flags, sort, and format (json|tsv|csv).",
	}, tagsListHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_aliases",
		Description: "List aliases for a note or across the vault. Optional file/path scope, total/verbose/active flags.",
	}, aliasesHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_links",
		Description: "List outgoing links from a note. Specify file or path; optional total flag.",
	}, linksHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_backlinks",
		Description: "List backlinks pointing to a note. Specify file or path; optional counts/total flags and format (json|tsv|csv).",
	}, backlinksHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_unresolved",
		Description: "List unresolved (broken) wikilinks in the vault. Optional total/counts/verbose flags and format (json|tsv|csv).",
	}, unresolvedHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_orphans",
		Description: "List notes with no incoming links. Optional total flag; all=true includes non-markdown files.",
	}, orphansHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_deadends",
		Description: "List notes with no outgoing links. Optional total flag; all=true includes non-markdown files.",
	}, deadendsHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_tag_query",
		Description: "Find notes by tag set algebra: all (AND), any (OR), none (NOT), optional path_prefix scope. Returns matching note paths. Use for microindex queries like all=[kind/clas, topic/auth].",
	}, tagQueryHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_outline",
		Description: "Show the heading outline of a note. Specify file or path; optional total flag and format (tree|md|json).",
	}, outlineHandler)
}
