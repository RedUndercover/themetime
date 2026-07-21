package buildinfo

import "testing"

func TestString(t *testing.T) {
	originalVersion, originalCommit := Version, Commit
	t.Cleanup(func() {
		Version, Commit = originalVersion, originalCommit
	})

	Version, Commit = "1.2.3", "abc1234"
	if got, want := String(), "1.2.3 (abc1234)"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}

	Commit = "unknown"
	if got, want := String(), "1.2.3"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}
