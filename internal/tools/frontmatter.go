package tools

import "strings"

// splitFrontmatter separates a note into its YAML frontmatter block (without
// the --- fences) and the body. Notes without a frontmatter block return
// ("", content).
func splitFrontmatter(content string) (frontmatter, body string) {
	rest, ok := strings.CutPrefix(content, "---\n")
	if !ok {
		return "", content
	}
	if fm, after, found := strings.Cut(rest, "\n---\n"); found {
		return fm, after
	}
	// frontmatter block closed at EOF without trailing newline
	if fm, found := strings.CutSuffix(rest, "\n---"); found {
		return fm, ""
	}
	return "", content
}

// sliceNote narrows note content to its frontmatter or body when requested.
func sliceNote(content string, frontmatterOnly, bodyOnly bool) string {
	if !frontmatterOnly && !bodyOnly {
		return content
	}
	fm, body := splitFrontmatter(content)
	if frontmatterOnly {
		return fm
	}
	return body
}
