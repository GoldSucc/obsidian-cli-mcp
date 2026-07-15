package tools

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gopkg.in/yaml.v3"
)

type PropFilter struct {
	Key   string `json:"key" jsonschema:"frontmatter property name"`
	Op    string `json:"op,omitempty" jsonschema:"comparison operator: eq (default), ne, contains, lt, gt, exists, missing"`
	Value string `json:"value,omitempty" jsonschema:"comparison value; ignored for exists/missing"`
}

type FilesQueryInput struct {
	Vault  string       `json:"vault,omitempty" jsonschema:"target vault name; defaults to OBSIDIAN_DEFAULT_VAULT or most recent"`
	Folder string       `json:"folder,omitempty" jsonschema:"folder path from vault root to scan; defaults to whole vault"`
	Where  []PropFilter `json:"where" jsonschema:"frontmatter property filters; a note must satisfy all of them"`
	Select []string     `json:"select,omitempty" jsonschema:"frontmatter properties to include next to each path"`
	Limit  int          `json:"limit,omitempty" jsonschema:"maximum number of notes to return"`
}

// propValues normalizes a frontmatter value to comparable strings: scalars
// become a single element, lists one element each.
func propValues(v any) []string {
	switch t := v.(type) {
	case nil:
		return nil
	case []any:
		out := make([]string, 0, len(t))
		for _, e := range t {
			out = append(out, strings.TrimSpace(fmt.Sprint(e)))
		}
		return out
	default:
		return []string{strings.TrimSpace(fmt.Sprint(t))}
	}
}

// compareOrdered compares numerically when both sides parse as numbers,
// falling back to lexicographic comparison (which orders ISO dates correctly).
func compareOrdered(a, b string) int {
	fa, errA := strconv.ParseFloat(a, 64)
	fb, errB := strconv.ParseFloat(b, 64)
	if errA == nil && errB == nil {
		switch {
		case fa < fb:
			return -1
		case fa > fb:
			return 1
		}
		return 0
	}
	return strings.Compare(a, b)
}

func matchFilter(props map[string]any, f PropFilter) bool {
	raw, present := props[f.Key]
	switch f.Op {
	case "exists":
		return present
	case "missing":
		return !present
	}
	if !present {
		return false
	}
	values := propValues(raw)
	switch f.Op {
	case "", "eq":
		return slices.Contains(values, f.Value)
	case "ne":
		return !slices.Contains(values, f.Value)
	case "contains":
		for _, v := range values {
			if strings.Contains(v, f.Value) {
				return true
			}
		}
		return false
	case "lt":
		for _, v := range values {
			if compareOrdered(v, f.Value) < 0 {
				return true
			}
		}
		return false
	case "gt":
		for _, v := range values {
			if compareOrdered(v, f.Value) > 0 {
				return true
			}
		}
		return false
	}
	return false
}

func noteProps(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	fm, _ := splitFrontmatter(string(data))
	if fm == "" {
		return map[string]any{}, nil
	}
	props := map[string]any{}
	if err := yaml.Unmarshal([]byte(fm), &props); err != nil {
		return map[string]any{}, nil
	}
	return props, nil
}

func filesQueryHandler(ctx context.Context, _ *mcp.CallToolRequest, in FilesQueryInput) (*mcp.CallToolResult, TextOutput, error) {
	if len(in.Where) == 0 {
		return nil, TextOutput{}, fmt.Errorf("where is required — use obsidian_files_list for unfiltered listings")
	}
	for _, f := range in.Where {
		if f.Key == "" {
			return nil, TextOutput{}, fmt.Errorf("every filter needs a key")
		}
	}
	root, err := exec.VaultPath(ctx, in.Vault)
	if err != nil {
		return nil, TextOutput{}, err
	}
	scan := root
	if in.Folder != "" {
		scan, err = exec.SecureJoin(root, in.Folder)
		if err != nil {
			return nil, TextOutput{}, err
		}
	}
	type row struct {
		rel   string
		props map[string]any
	}
	var rows []row
	walkErr := filepath.WalkDir(scan, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && path != scan {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		props, pErr := noteProps(path)
		if pErr != nil {
			return nil
		}
		for _, f := range in.Where {
			if !matchFilter(props, f) {
				return nil
			}
		}
		rel, rErr := filepath.Rel(root, path)
		if rErr != nil {
			return nil
		}
		rows = append(rows, row{rel: filepath.ToSlash(rel), props: props})
		return nil
	})
	if walkErr != nil {
		return nil, TextOutput{}, walkErr
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].rel < rows[j].rel })
	total := len(rows)
	if in.Limit > 0 && total > in.Limit {
		rows = rows[:in.Limit]
	}
	var b strings.Builder
	unit := "notes"
	if total == 1 {
		unit = "note"
	}
	fmt.Fprintf(&b, "%d %s", total, unit)
	if in.Limit > 0 && total > in.Limit {
		fmt.Fprintf(&b, " (showing %d)", in.Limit)
	}
	for _, r := range rows {
		b.WriteString("\n")
		b.WriteString(r.rel)
		for _, key := range in.Select {
			vals := propValues(r.props[key])
			b.WriteString("\t")
			b.WriteString(key + "=" + strings.Join(vals, ","))
		}
	}
	return nil, TextOutput{Content: b.String()}, nil
}

func RegisterQuery(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_files_query",
		Description: "Find notes by frontmatter property filters (eq, ne, contains, lt, gt, exists, missing). Optional folder scope, select to include property values, limit. Example: where=[{key: kind, value: clas}, {key: last-updated, op: lt, value: 2026-01-01}].",
	}, filesQueryHandler)
}
