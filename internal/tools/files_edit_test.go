package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/GoldSucc/obsidian-cli-mcp/internal/exec"
)

// These tests drive the real handlers against a live Obsidian vault via the
// `obsidian` CLI. They require the Obsidian app to be running. Run explicitly:
//   go test ./internal/tools -run TestEditReplace -v
const testNotePath = "Home/zz-edit-replace-test.md"

func readNote(t *testing.T, path string) string {
	t.Helper()
	out, err := exec.Run(context.Background(), exec.Args{Command: "read", Params: map[string]string{"path": path}})
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return out
}

func TestEditReplace(t *testing.T) {
	ctx := context.Background()
	defer exec.Run(ctx, exec.Args{Command: "delete", Params: map[string]string{"path": testNotePath}, Flags: []string{"permanent"}})

	// seed
	if _, _, err := createHandler(ctx, nil, CreateInput{Path: testNotePath, Content: "alpha\nbeta\ngamma"}); err != nil {
		t.Fatalf("seed create: %v", err)
	}

	// edit: unique single replace
	if _, out, err := editHandler(ctx, nil, EditInput{FileTarget: FileTarget{Path: testNotePath}, OldString: "beta", NewString: "BETA"}); err != nil {
		t.Fatalf("edit: %v", err)
	} else {
		t.Logf("edit -> %s", out.Content)
	}
	if got := readNote(t, testNotePath); !strings.Contains(got, "BETA") || strings.Contains(got, "\nbeta\n") {
		t.Fatalf("edit not applied, got: %q", got)
	}

	// edit: not found
	if _, _, err := editHandler(ctx, nil, EditInput{FileTarget: FileTarget{Path: testNotePath}, OldString: "nope", NewString: "x"}); err == nil {
		t.Fatal("expected not-found error")
	}

	// edit: identical strings rejected
	if _, _, err := editHandler(ctx, nil, EditInput{FileTarget: FileTarget{Path: testNotePath}, OldString: "x", NewString: "x"}); err == nil {
		t.Fatal("expected identical-string error")
	}

	// edit: non-unique without replace_all rejected
	if _, _, err := editHandler(ctx, nil, EditInput{FileTarget: FileTarget{Path: testNotePath}, OldString: "a", NewString: "A"}); err == nil {
		t.Fatal("expected non-unique error")
	}

	// edit: replace_all
	if _, _, err := editHandler(ctx, nil, EditInput{FileTarget: FileTarget{Path: testNotePath}, OldString: "a", NewString: "@", ReplaceAll: true}); err != nil {
		t.Fatalf("edit replace_all: %v", err)
	}
	if got := readNote(t, testNotePath); strings.Contains(got, "a") {
		t.Fatalf("replace_all left 'a': %q", got)
	}

	// replace: full overwrite
	if _, _, err := replaceHandler(ctx, nil, ReplaceInput{FileTarget: FileTarget{Path: testNotePath}, Content: "fresh body\nline 2"}); err != nil {
		t.Fatalf("replace: %v", err)
	}
	if got := readNote(t, testNotePath); !strings.HasPrefix(got, "fresh body") {
		t.Fatalf("replace not applied: %q", got)
	}
}
