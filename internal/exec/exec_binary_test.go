package exec

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The Obsidian app bundle ships the CLI as `obsidian-cli` next to the app
// binary `Obsidian`. On macOS's case-insensitive filesystem a PATH lookup for
// `obsidian` resolves to the app, which boots a full Obsidian instance instead
// of answering the command. Run must pick the CLI even when the app binary is
// sitting in the same directory.
func TestRunPrefersCLIOverAppBinary(t *testing.T) {
	dir := t.TempDir()
	write := func(name, marker string) {
		script := "#!/bin/sh\necho " + marker + "\n"
		if err := os.WriteFile(filepath.Join(dir, name), []byte(script), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write("Obsidian", "APP")
	write("obsidian-cli", "CLI")
	t.Setenv("PATH", dir)

	out, err := Run(context.Background(), Args{Command: "version"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := strings.TrimSpace(out); got != "CLI" {
		t.Fatalf("ran the wrong binary: got %q, want %q", got, "CLI")
	}
}
