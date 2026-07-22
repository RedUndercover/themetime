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

func TestApplyVideoWallpaperResetsPluginInSeparateCall(t *testing.T) {
	runner := &recordingRunner{}
	err := applyWallpaperScript(context.Background(), runner, SmartVideoWallpaperPlugin, "1", map[string]string{
		"VideoUrls": `[{"filename":"/wall/video.mp4"}]`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(runner.calls) != 2 {
		t.Fatalf("qdbus calls = %d, want reset and apply calls", len(runner.calls))
	}
	resetScript := runner.calls[0][len(runner.calls[0])-1]
	if !strings.Contains(resetScript, `desktop.wallpaperPlugin = "org.kde.image"`) {
		t.Fatalf("first call does not reset the wallpaper plugin:\n%s", resetScript)
	}
	if !strings.Contains(resetScript, `String(desktop.screen) !== targetScreen`) {
		t.Fatalf("reset call does not preserve the screen target:\n%s", resetScript)
	}
	applyScript := runner.calls[1][len(runner.calls[1])-1]
	if !strings.Contains(applyScript, SmartVideoWallpaperPlugin) {
		t.Fatalf("second call does not apply the video plugin:\n%s", applyScript)
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
	if values["ResumeLastVideo"] != "false" {
		t.Fatalf("ResumeLastVideo = %q, want false to avoid resuming a stale playlist entry", values["ResumeLastVideo"])
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

type recordingRunner struct {
	calls [][]string
}

func (r *recordingRunner) LookPath(name string) (string, error) {
	return "/usr/bin/" + name, nil
}

func (r *recordingRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	call := append([]string{name}, args...)
	r.calls = append(r.calls, call)
	return "", nil
}
