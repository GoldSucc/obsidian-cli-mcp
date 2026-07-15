package exec

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCLISafe(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"plain text", true},
		{"multi\nline\ttext", true},
		{`C:\notes\x`, false},
		{`\newcommand{}`, false},
		{"regex `\\d+`", false},
		{"", true},
	}
	for _, c := range cases {
		if got := CLISafe(c.in); got != c.want {
			t.Errorf("CLISafe(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// The CLI decodes literal \n and \t with a blind replace. EncodeMultiline is
// only lossless for CLI-safe content — this pins the contract down.
func TestEncodeMultilineRoundTrip(t *testing.T) {
	decode := func(s string) string {
		s = strings.ReplaceAll(s, `\n`, "\n")
		return strings.ReplaceAll(s, `\t`, "\t")
	}
	safe := []string{"a\nb\tc", "no specials", "trailing\n"}
	for _, s := range safe {
		if !CLISafe(s) {
			t.Fatalf("test input %q should be CLI-safe", s)
		}
		if got := decode(EncodeMultiline(s)); got != s {
			t.Errorf("round trip of %q = %q", s, got)
		}
	}
	// Documents WHY unsafe content must bypass the CLI: the round trip corrupts it.
	unsafe := `C:\notes\x`
	if got := decode(EncodeMultiline(unsafe)); got == unsafe {
		t.Errorf("expected lossy round trip for %q — if this now passes, the CLI fallback can be removed", unsafe)
	}
}

func TestStdoutError(t *testing.T) {
	cases := []struct {
		out  string
		want string
	}{
		{`Error: File "x.md" not found.`, `Error: File "x.md" not found.`},
		{"Error: something\n", "Error: something"},
		{"Error: outage at 09:00\nRoot cause: DNS\n", ""}, // multi-line = note content
		{"normal output", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := stdoutError(c.out); got != c.want {
			t.Errorf("stdoutError(%q) = %q, want %q", c.out, got, c.want)
		}
	}
}

func TestSecureJoin(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "vault")
	ok := []string{"note.md", "folder/note.md", "a/b/../c.md"}
	for _, rel := range ok {
		if _, err := SecureJoin(root, rel); err != nil {
			t.Errorf("SecureJoin(%q) unexpected error: %v", rel, err)
		}
	}
	bad := []string{"../outside.md", "a/../../outside.md", "../../tmp/evil"}
	for _, rel := range bad {
		if _, err := SecureJoin(root, rel); err == nil {
			t.Errorf("SecureJoin(%q) should have been rejected", rel)
		}
	}
}
