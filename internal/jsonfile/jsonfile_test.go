package jsonfile

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fixture struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func TestWriteAtomicRoundTripAndReplace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "state.json")
	if err := WriteAtomic(path, fixture{Name: "first", Count: 1}); err != nil {
		t.Fatal(err)
	}
	if err := WriteAtomic(path, fixture{Name: "second", Count: 2}); err != nil {
		t.Fatal(err)
	}

	var got fixture
	if err := Read(path, &got); err != nil {
		t.Fatal(err)
	}
	if got != (fixture{Name: "second", Count: 2}) {
		t.Fatalf("decoded fixture = %#v", got)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 600", info.Mode().Perm())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(data), "\n") || !strings.Contains(string(data), "\n  \"name\"") {
		t.Fatalf("JSON is not indented and newline-terminated: %q", data)
	}
	if matches, err := filepath.Glob(filepath.Join(filepath.Dir(path), ".state.json.*.tmp")); err != nil || len(matches) != 0 {
		t.Fatalf("temporary files after success = %v, err = %v", matches, err)
	}
}

func TestReadReportsMalformedPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.json")
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	var got fixture
	err := Read(path, &got)
	if err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("error = %v, want path %q", err, path)
	}
}

func TestWriteAtomicCleansUpAfterMarshalFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	if err := WriteAtomic(path, func() {}); err == nil {
		t.Fatal("expected marshal failure")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("files after marshal failure = %v", entries)
	}
}
