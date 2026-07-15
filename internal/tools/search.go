package tools

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

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
	Total  bool   `json:"total,omitempty" jsonschema:"return only the note and match counts"`
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
	// json is a structured passthrough; only reshape the human-readable text view.
	if in.Total {
		out = summarizeSearchContext(out)
	} else if in.Format != "json" {
		out = groupSearchContext(out)
	}
	return nil, TextOutput{Content: out}, nil
}

// summarizeSearchContext reduces `search:context` output to counts. The CLI
// has no total flag for this command, so the reduction happens here.
func summarizeSearchContext(raw string) string {
	notes := map[string]bool{}
	seen := map[string]bool{}
	for ln := range strings.SplitSeq(strings.TrimRight(raw, "\n"), "\n") {
		m := contextLineRe.FindStringSubmatch(ln)
		if m == nil {
			continue
		}
		if seen[ln] {
			continue
		}
		seen[ln] = true
		notes[m[1]] = true
	}
	return fmt.Sprintf("%d notes, %d matches", len(notes), len(seen))
}

// contextLineRe splits a `search:context` line into path, line number, and the
// matched text. The non-greedy path anchors on the first `:<digits>:` segment so
// colons in either the path or the content do not break the parse.
var contextLineRe = regexp.MustCompile(`^(.+?):(\d+):\s?(.*)$`)

// groupSearchContext collapses the per-match `search:context` output into one
// block per note: the note path, its match count, and the first matching line.
// Identical (path, line) entries are de-duplicated so the count is not inflated
// by the CLI emitting the same line more than once.
func groupSearchContext(raw string) string {
	type note struct {
		path    string
		first   string
		matches int
		seen    map[string]bool
	}
	order := []string{}
	notes := map[string]*note{}
	for ln := range strings.SplitSeq(strings.TrimRight(raw, "\n"), "\n") {
		m := contextLineRe.FindStringSubmatch(ln)
		if m == nil {
			continue
		}
		path, lineNo, content := m[1], m[2], m[3]
		n, ok := notes[path]
		if !ok {
			n = &note{path: path, seen: map[string]bool{}}
			notes[path] = n
			order = append(order, path)
		}
		key := lineNo + "\x00" + content
		if n.seen[key] {
			continue
		}
		n.seen[key] = true
		n.matches++
		if n.first == "" {
			n.first = "L" + lineNo + ": " + content
		}
	}
	var b strings.Builder
	for i, path := range order {
		n := notes[path]
		if i > 0 {
			b.WriteString("\n")
		}
		unit := "matches"
		if n.matches == 1 {
			unit = "match"
		}
		fmt.Fprintf(&b, "%s  (%d %s)\n  %s\n", n.path, n.matches, unit, n.first)
	}
	return b.String()
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
		Description: "Search vault and return a list of matching notes, each with its match count and the first matching line. Optional path, limit, case. Use format=json for the raw per-match list.",
	}, searchContextHandler)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_search_open",
		Description: "Open the Obsidian search panel, optionally prefilled with a query.",
	}, searchOpenHandler)
}
