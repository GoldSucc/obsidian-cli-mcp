package tools

import "testing"

func TestGroupSearchContext(t *testing.T) {
	raw := "" +
		"a/note.md:39: | `make test-docker` | All three suites in Docker |\n" +
		"a/note.md:39: | `make test-docker` | All three suites in Docker |\n" +
		"b/testing.md:16: Three test suites. Run via `make test-docker`.\n" +
		"b/testing.md:56: ## Docker (all suites)\n"

	got := groupSearchContext(raw)
	want := "" +
		"a/note.md  (1 match)\n" +
		"  L39: | `make test-docker` | All three suites in Docker |\n" +
		"\n" +
		"b/testing.md  (2 matches)\n" +
		"  L16: Three test suites. Run via `make test-docker`.\n"

	if got != want {
		t.Fatalf("grouped output mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestGroupSearchContextEmpty(t *testing.T) {
	if got := groupSearchContext(""); got != "" {
		t.Fatalf("expected empty output, got %q", got)
	}
}
