package main

import (
	"bytes"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/RedUndercover/themetime/internal/model"
)

func TestPreviewLocationDoesNotMutateConfig(t *testing.T) {
	cfg := model.DefaultConfig()
	original := cfg.Location
	location := model.Location{
		Label:     "Reykjavik",
		Latitude:  64.1466,
		Longitude: -21.9426,
		Timezone:  "Atlantic/Reykjavik",
	}
	now := time.Date(2026, 3, 20, 12, 0, 0, 0, time.UTC)

	preview, err := previewLocation(cfg, location, now)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Location != original {
		t.Fatalf("preview mutated config location: got %#v, want %#v", cfg.Location, original)
	}
	if preview.Today == "" {
		t.Fatal("preview Today is empty")
	}
	if len(preview.SolarEvents) != len(model.SolarTriggerKinds()) {
		t.Fatalf("solar events = %d, want %d", len(preview.SolarEvents), len(model.SolarTriggerKinds()))
	}
	if len(preview.Transitions) != len(cfg.Phases) {
		t.Fatalf("transitions = %d, want %d", len(preview.Transitions), len(cfg.Phases))
	}
}

func TestPreviewLocationRejectsUnknownTimezone(t *testing.T) {
	cfg := model.DefaultConfig()
	location := cfg.Location
	location.Timezone = "Mars/Olympus_Mons"

	if _, err := previewLocation(cfg, location, time.Now()); err == nil {
		t.Fatal("expected unknown timezone to fail")
	}
}

func TestThemeTimeIconIsPNG(t *testing.T) {
	data := themeTimeIconPNG()
	image, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if image.Bounds().Dx() != 64 || image.Bounds().Dy() != 64 {
		t.Fatalf("icon bounds = %v, want 64x64", image.Bounds())
	}
}

func TestNormalizeMediaPathsExpandsHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	cfg := model.DefaultConfig()
	cfg.Phases[0].Actions = []model.Action{{Type: model.ActionVideoWallpaper, Value: "~/.wallpapers/day.mp4"}}

	got := normalizeMediaPaths(cfg)
	want := filepath.Join(home, ".wallpapers", "day.mp4")
	if got.Phases[0].Actions[0].Value != want {
		t.Fatalf("normalized path = %q, want %q", got.Phases[0].Actions[0].Value, want)
	}
	if cfg.Phases[0].Actions[0].Value != "~/.wallpapers/day.mp4" {
		t.Fatal("normalization mutated its input config")
	}
}

func TestUIStateExposesEachTriggerDefinitionOnce(t *testing.T) {
	app := &App{cfg: model.DefaultConfig()}
	state, err := app.state()
	if err != nil {
		t.Fatal(err)
	}
	definitions := model.TriggerDefinitions()
	if len(state.TriggerOptions) != len(definitions) {
		t.Fatalf("trigger options = %d, want %d", len(state.TriggerOptions), len(definitions))
	}
	seen := map[model.TriggerKind]bool{}
	for i, option := range state.TriggerOptions {
		if seen[option.Kind] {
			t.Fatalf("duplicate trigger option %q", option.Kind)
		}
		seen[option.Kind] = true
		if option != definitions[i] {
			t.Fatalf("trigger option %d = %#v, want %#v", i, option, definitions[i])
		}
	}
}
