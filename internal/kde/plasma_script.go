package kde

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const SmartVideoWallpaperPlugin = "luisbocanegra.smart.video.wallpaper.reborn"

type VideoItem struct {
	Filename                string  `json:"filename"`
	Enabled                 bool    `json:"enabled"`
	Duration                int     `json:"duration"`
	CustomDuration          bool    `json:"customDuration"`
	PlaybackRate            float64 `json:"playbackRate"`
	AlternativePlaybackRate float64 `json:"alternativePlaybackRate"`
	Loop                    bool    `json:"loop"`
}

func applyWallpaperScript(ctx context.Context, r Runner, plugin string, screen string, values map[string]string) error {
	script, err := wallpaperScript(plugin, screen, values)
	if err != nil {
		return err
	}
	qdbus, ok := bestCommand(r, "qdbus6", "qdbus")
	if !ok {
		return fmt.Errorf("qdbus6 is required to configure per-screen wallpapers")
	}
	_, err = r.Run(ctx, qdbus, "org.kde.plasmashell", "/PlasmaShell", "org.kde.PlasmaShell.evaluateScript", script)
	return err
}

func wallpaperScript(plugin string, screen string, values map[string]string) (string, error) {
	if plugin == "" {
		return "", fmt.Errorf("wallpaper plugin is required")
	}
	var b strings.Builder
	b.WriteString("var targetScreen = ")
	b.WriteString(jsString(screen))
	b.WriteString(";\n")
	b.WriteString("desktops().forEach(function(desktop) {\n")
	b.WriteString("  if (targetScreen !== \"\" && String(desktop.screen) !== targetScreen) { return; }\n")
	b.WriteString("  desktop.wallpaperPlugin = ")
	b.WriteString(jsString(plugin))
	b.WriteString(";\n")
	b.WriteString("  desktop.currentConfigGroup = [\"Wallpaper\", ")
	b.WriteString(jsString(plugin))
	b.WriteString(", \"General\"];\n")
	for key, value := range values {
		b.WriteString("  desktop.writeConfig(")
		b.WriteString(jsString(key))
		b.WriteString(", ")
		b.WriteString(jsString(value))
		b.WriteString(");\n")
	}
	b.WriteString("  desktop.reloadConfig();\n")
	b.WriteString("});\n")
	return b.String(), nil
}

func videoWallpaperValues(videoPath string, extra map[string]string) (map[string]string, error) {
	item := VideoItem{
		Filename:                videoPath,
		Enabled:                 true,
		Duration:                0,
		CustomDuration:          false,
		PlaybackRate:            1,
		AlternativePlaybackRate: 0.5,
		Loop:                    true,
	}
	if extra["duration"] != "" {
		fmt.Sscanf(extra["duration"], "%d", &item.Duration)
		if item.Duration > 0 {
			item.CustomDuration = true
		}
	}
	data, err := json.Marshal([]VideoItem{item})
	if err != nil {
		return nil, err
	}
	values := map[string]string{
		"VideoUrls":                   string(data),
		"ChangeWallpaperMode":         "0",
		"RandomMode":                  "false",
		"ResumeLastVideo":             "true",
		"MuteMode":                    valueOrDefault(extra["muteMode"], "5"),
		"FillMode":                    valueOrDefault(extra["fillMode"], "2"),
		"PauseMode":                   valueOrDefault(extra["pauseMode"], "3"),
		"BatteryPausesVideo":          valueOrDefault(extra["batteryPausesVideo"], "true"),
		"ScreenOffPausesVideo":        valueOrDefault(extra["screenOffPausesVideo"], "true"),
		"CrossfadeEnabled":            valueOrDefault(extra["crossfadeEnabled"], "false"),
		"CheckWindowsActiveScreen":    valueOrDefault(extra["checkWindowsActiveScreen"], "true"),
		"AlternativePlaybackRateMode": valueOrDefault(extra["alternativePlaybackRateMode"], "3"),
	}
	if extra["volume"] != "" {
		values["Volume"] = extra["volume"]
	}
	if extra["blurMode"] != "" {
		values["BlurMode"] = extra["blurMode"]
	}
	return values, nil
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func jsString(value string) string {
	data, _ := json.Marshal(value)
	return string(data)
}
