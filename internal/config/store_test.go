package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/RedUndercover/themetime/internal/model"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	want := model.DefaultConfig()
	want.Version = 0
	if err := Save(path, want); err != nil {
		t.Fatal(err)
	}
	want.Version = model.CurrentConfigVersion
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("loaded config = %#v, want %#v", got, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 600", info.Mode().Perm())
	}
}

func TestSaveRejectsInvalidConfigWithoutCreatingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	cfg := model.DefaultConfig()
	cfg.Location.Latitude = 100
	if err := Save(path, cfg); err == nil {
		t.Fatal("expected validation error")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("stat error = %v, want not-exist", err)
	}
}

func TestLoadRejectsMalformedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("load error = %v, want malformed JSON error containing %q", err, path)
	}
}
