package kde

import (
	"slices"
	"testing"
)

func TestWallpaperExtensionListsAreCompleteAndIndependent(t *testing.T) {
	static := StaticWallpaperExtensions()
	video := VideoWallpaperExtensions()
	for _, extension := range []string{".avif", ".jpg", ".png", ".svg", ".webp"} {
		if !slices.Contains(static, extension) {
			t.Fatalf("static extensions %v do not include %q", static, extension)
		}
	}
	for _, extension := range []string{".mkv", ".mp4", ".webm"} {
		if !slices.Contains(video, extension) {
			t.Fatalf("video extensions %v do not include %q", video, extension)
		}
	}
	static[0] = ".changed"
	if StaticWallpaperExtensions()[0] == ".changed" {
		t.Fatal("StaticWallpaperExtensions returned mutable shared state")
	}
}
