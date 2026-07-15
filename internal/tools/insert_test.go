package tools

import (
	"strings"
	"testing"
)

const insertFixture = `---
tags: [x]
---
# Title

intro line

## Gotchas

- first gotcha

## Links

- [[other]]`

func TestFindHeading(t *testing.T) {
	lines := strings.Split(insertFixture, "\n")
	idx, level := findHeading(lines, "Gotchas")
	if idx < 0 || level != 2 {
		t.Fatalf("findHeading(Gotchas) = %d, %d", idx, level)
	}
	if lines[idx] != "## Gotchas" {
		t.Fatalf("matched wrong line: %q", lines[idx])
	}
	// leading # marks in the query are ignored
	if i2, _ := findHeading(lines, "## Gotchas"); i2 != idx {
		t.Fatalf("hash-prefixed query matched %d, want %d", i2, idx)
	}
	if i3, _ := findHeading(lines, "Missing"); i3 != -1 {
		t.Fatalf("expected no match, got %d", i3)
	}
}

func TestInsertLinesEnd(t *testing.T) {
	lines := strings.Split(insertFixture, "\n")
	idx, level := findHeading(lines, "Gotchas")
	updated := strings.Join(insertLines(lines, idx, level, "- second gotcha", ""), "\n")
	want := "- first gotcha\n- second gotcha\n\n## Links"
	if !strings.Contains(updated, want) {
		t.Fatalf("end insert misplaced:\n%s", updated)
	}
}

func TestInsertLinesStart(t *testing.T) {
	lines := strings.Split(insertFixture, "\n")
	idx, level := findHeading(lines, "Gotchas")
	updated := strings.Join(insertLines(lines, idx, level, "- zero gotcha", "start"), "\n")
	want := "## Gotchas\n- zero gotcha\n\n- first gotcha"
	if !strings.Contains(updated, want) {
		t.Fatalf("start insert misplaced:\n%s", updated)
	}
}

func TestInsertLinesLastSection(t *testing.T) {
	lines := strings.Split(insertFixture, "\n")
	idx, level := findHeading(lines, "Links")
	updated := strings.Join(insertLines(lines, idx, level, "- [[new]]", ""), "\n")
	if !strings.HasSuffix(updated, "- [[other]]\n- [[new]]") {
		t.Fatalf("last-section insert misplaced:\n%s", updated)
	}
}
