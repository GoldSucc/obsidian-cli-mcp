package tools

import "testing"

func TestSplitFrontmatter(t *testing.T) {
	fm, body := splitFrontmatter("---\nkind: clas\n---\nbody line\n")
	if fm != "kind: clas" {
		t.Errorf("frontmatter = %q", fm)
	}
	if body != "body line\n" {
		t.Errorf("body = %q", body)
	}

	fm, body = splitFrontmatter("no frontmatter here")
	if fm != "" || body != "no frontmatter here" {
		t.Errorf("plain note mishandled: %q / %q", fm, body)
	}

	// frontmatter-only note, fence closed at EOF
	fm, body = splitFrontmatter("---\nkind: clas\n---")
	if fm != "kind: clas" || body != "" {
		t.Errorf("EOF fence mishandled: %q / %q", fm, body)
	}

	// a --- ruler later in the body is not a frontmatter fence
	fm, _ = splitFrontmatter("body first\n---\nnot frontmatter\n---\n")
	if fm != "" {
		t.Errorf("ruler misread as frontmatter: %q", fm)
	}
}

func TestSliceNote(t *testing.T) {
	note := "---\nkind: clas\n---\nbody\n"
	if got := sliceNote(note, true, false); got != "kind: clas" {
		t.Errorf("frontmatter_only = %q", got)
	}
	if got := sliceNote(note, false, true); got != "body\n" {
		t.Errorf("body_only = %q", got)
	}
	if got := sliceNote(note, false, false); got != note {
		t.Errorf("passthrough = %q", got)
	}
}

func TestMatchFilter(t *testing.T) {
	props := map[string]any{
		"kind":         "clas",
		"tags":         []any{"topic/auth", "kind/clas"},
		"last-updated": "2026-03-01",
		"count":        7,
	}
	cases := []struct {
		f    PropFilter
		want bool
	}{
		{PropFilter{Key: "kind", Value: "clas"}, true},
		{PropFilter{Key: "kind", Op: "eq", Value: "intf"}, false},
		{PropFilter{Key: "tags", Value: "topic/auth"}, true},
		{PropFilter{Key: "tags", Op: "contains", Value: "auth"}, true},
		{PropFilter{Key: "kind", Op: "ne", Value: "intf"}, true},
		{PropFilter{Key: "last-updated", Op: "lt", Value: "2026-07-01"}, true},
		{PropFilter{Key: "last-updated", Op: "gt", Value: "2026-07-01"}, false},
		{PropFilter{Key: "count", Op: "gt", Value: "10"}, false},
		{PropFilter{Key: "count", Op: "lt", Value: "10"}, true},
		{PropFilter{Key: "kind", Op: "exists"}, true},
		{PropFilter{Key: "owner", Op: "exists"}, false},
		{PropFilter{Key: "owner", Op: "missing"}, true},
		{PropFilter{Key: "owner", Value: "x"}, false},
	}
	for _, c := range cases {
		if got := matchFilter(props, c.f); got != c.want {
			t.Errorf("matchFilter(%+v) = %v, want %v", c.f, got, c.want)
		}
	}
}
