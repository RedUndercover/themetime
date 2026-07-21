package systemd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDaemonBinaryFromWailsPrefersSiblingDaemon(t *testing.T) {
	dir := t.TempDir()
	gui := filepath.Join(dir, "themetime-wails")
	daemon := filepath.Join(dir, "themetime")
	if err := os.WriteFile(gui, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(daemon, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := resolveDaemonBinaryFromExecutable("", gui)
	if err != nil {
		t.Fatal(err)
	}
	if got != daemon {
		t.Fatalf("daemon path = %q, want %q", got, daemon)
	}
}

func TestResolveDaemonBinaryRejectsGoRunExecutable(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	dir := filepath.Join(t.TempDir(), "go-build123", "b001", "exe")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(dir, "themetime")
	if err := os.WriteFile(executable, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	if _, err := resolveDaemonBinaryFromExecutable("", executable); err == nil {
		t.Fatal("expected temporary executable to be rejected")
	}
}

func TestSystemdQuotePathEscapesSpacesAndPercents(t *testing.T) {
	got := systemdQuotePath("/opt/Theme Time/%bin/themetime")
	want := `"/opt/Theme Time/%%bin/themetime"`
	if got != want {
		t.Fatalf("quoted path = %q, want %q", got, want)
	}
}
