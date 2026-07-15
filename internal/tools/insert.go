package tools

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type InsertInput struct {
	FileTarget
	Heading  string `json:"heading" jsonschema:"heading text to target, with or without leading # marks; matched against the first heading whose text equals it"`
	Content  string `json:"content" jsonschema:"content to insert; real newlines accepted"`
	Position string `json:"position,omitempty" jsonschema:"placement within the heading's section: start (directly under the heading) or end (after the section's last content line, before the next same-or-higher-level heading); default end"`
}

var headingLineRe = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s*$`)

// findHeading returns the index and level of the first line whose heading
// text equals want (leading # marks in want are ignored).
func findHeading(lines []string, want string) (idx, level int) {
	want = strings.TrimSpace(strings.TrimLeft(want, "#"))
	for i, line := range lines {
		m := headingLineRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		if strings.TrimSpace(m[2]) == want {
			return i, len(m[1])
		}
	}
	return -1, 0
}

// sectionEnd returns the index one past the last line belonging to the
// section opened by the heading at start: the next heading of the same or a
// higher level closes it.
func sectionEnd(lines []string, start, level int) int {
	for i := start + 1; i < len(lines); i++ {
		if m := headingLineRe.FindStringSubmatch(lines[i]); m != nil && len(m[1]) <= level {
			return i
		}
	}
	return len(lines)
}

// insertLines computes the updated note lines with content placed inside the
// section of the matched heading. Position end skips trailing blank lines so
// the insert lands after the section's last content line.
func insertLines(lines []string, headingIdx, level int, content, position string) []string {
	end := sectionEnd(lines, headingIdx, level)
	at := end
	if position == "start" {
		at = headingIdx + 1
	} else {
		for at > headingIdx+1 && strings.TrimSpace(lines[at-1]) == "" {
			at--
		}
	}
	inserted := strings.Split(content, "\n")
	updated := make([]string, 0, len(lines)+len(inserted))
	updated = append(updated, lines[:at]...)
	updated = append(updated, inserted...)
	updated = append(updated, lines[at:]...)
	return updated
}

func listHeadings(lines []string) []string {
	var out []string
	for _, line := range lines {
		if m := headingLineRe.FindStringSubmatch(line); m != nil {
			out = append(out, m[0])
		}
	}
	return out
}

func insertHandler(ctx context.Context, _ *mcp.CallToolRequest, in InsertInput) (*mcp.CallToolResult, TextOutput, error) {
	if in.Heading == "" {
		return nil, TextOutput{}, fmt.Errorf("heading is required")
	}
	if in.Content == "" {
		return nil, TextOutput{}, fmt.Errorf("content is required")
	}
	if in.Position != "" && in.Position != "start" && in.Position != "end" {
		return nil, TextOutput{}, fmt.Errorf("position must be start or end")
	}
	path, err := resolvePath(ctx, in.Vault, in.FileTarget)
	if err != nil {
		return nil, TextOutput{}, err
	}
	content, err := exec.Run(ctx, exec.Args{Command: "read", Vault: in.Vault, Params: map[string]string{"path": path}})
	if err != nil {
		return nil, TextOutput{}, err
	}
	lines := strings.Split(content, "\n")
	idx, level := findHeading(lines, in.Heading)
	if idx < 0 {
		available := listHeadings(lines)
		if len(available) == 0 {
			return nil, TextOutput{}, fmt.Errorf("no headings in %s", path)
		}
		return nil, TextOutput{}, fmt.Errorf("heading %q not found in %s; available: %s", in.Heading, path, strings.Join(available, " | "))
	}
	updated := insertLines(lines, idx, level, in.Content, in.Position)
	if _, err := writeContent(ctx, in.Vault, path, strings.Join(updated, "\n"), true); err != nil {
		return nil, TextOutput{}, err
	}
	return nil, TextOutput{Content: fmt.Sprintf("Inserted under %q in %s", strings.TrimSpace(lines[idx]), path)}, nil
}

func RegisterInsert(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "obsidian_insert",
		Description: "Insert content inside a note section identified by its heading text. position=start puts it directly under the heading, position=end (default) after the section's last content line. Target by file (wikilink-style) or path (exact).",
	}, insertHandler)
}
