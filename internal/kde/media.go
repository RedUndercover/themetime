package kde

import "slices"

var (
	staticMediaExtensions = []string{".avif", ".bmp", ".jpeg", ".jpg", ".png", ".svg", ".webp"}
	videoMediaExtensions  = []string{".avi", ".m4v", ".mkv", ".mov", ".mp4", ".webm"}
)

func StaticWallpaperExtensions() []string {
	return slices.Clone(staticMediaExtensions)
}

func VideoWallpaperExtensions() []string {
	return slices.Clone(videoMediaExtensions)
}
