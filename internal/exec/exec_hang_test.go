package exec

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// A CLI process that leaves a child holding the stdout/stderr pipes must not
// block Run past the timeout: exec.CommandContext only kills the direct
// child, and without cmd.WaitDelay, Run keeps waiting for the pipes to drain
// until every descendant exits — the "forever-running MCP call" failure mode.
func TestRunReturnsDespiteLingeringChild(t *testing.T) {
	dir := t.TempDir()
	stub := filepath.Join(dir, Binary)
	script := "#!/bin/sh\n/bin/sleep 20 &\nexec /bin/sleep 20\n"
	if err := os.WriteFile(stub, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("OBSIDIAN_MCP_TIMEOUT", "1")

	start := time.Now()
	_, err := Run(context.Background(), Args{Command: "version"})
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if elapsed > 8*time.Second {
		t.Fatalf("Run blocked %s waiting on lingering child; want return shortly after the 1s timeout", elapsed)
	}
	t.Logf("returned after %s with: %v", elapsed, err)
}
