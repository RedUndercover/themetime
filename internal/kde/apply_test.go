package kde

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWallpaperScriptEscapesValues(t *testing.T) {
	script, err := wallpaperScript("org.kde.image", "1", map[string]string{
		"Image": `file:///tmp/a "quoted".png`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(script, `"file:///tmp/a \"quoted\".png"`) {
		t.Fatalf("script did not escape quoted value:\n%s", script)
	}
	if !strings.Contains(script, `String(desktop.screen) !== targetScreen`) {
		t.Fatalf("script missing screen guard:\n%s", script)
	}
}

func TestVideoWallpaperValues(t *testing.T) {
	values, err := videoWallpaperValues("/wall/video.mp4", map[string]string{"volume": "0.2"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(values["VideoUrls"], "/wall/video.mp4") {
		t.Fatalf("VideoUrls = %s", values["VideoUrls"])
	}
	if values["Volume"] != "0.2" {
		t.Fatalf("Volume = %q", values["Volume"])
	}
}

func TestExpandUserPathForMedia(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	got := expandUserPath("~/.wallpapers/day.mp4")
	want := filepath.Join(home, ".wallpapers", "day.mp4")
	if got != want {
		t.Fatalf("expanded path = %q, want %q", got, want)
	}
}

func TestDiscoverDoesNotPanicWithMissingCommands(t *testing.T) {
	inv := Discover(context.Background(), fakeRunner{})
	if inv.Commands == nil {
		t.Fatal("commands map is nil")
	}
}

type fakeRunner struct{}

func (fakeRunner) LookPath(name string) (string, error) {
	return "", context.Canceled
}

func (fakeRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	return "", context.Canceled
}
