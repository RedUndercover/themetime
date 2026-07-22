package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/RedUndercover/themetime/internal/kde"
	"github.com/RedUndercover/themetime/internal/model"
)

var hexColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func normalizeMediaPaths(cfg model.Config) model.Config {
	out := cfg
	out.Phases = append([]model.Phase(nil), cfg.Phases...)
	for phaseIndex := range out.Phases {
		out.Phases[phaseIndex].Actions = append([]model.Action(nil), cfg.Phases[phaseIndex].Actions...)
		for actionIndex := range out.Phases[phaseIndex].Actions {
			action := &out.Phases[phaseIndex].Actions[actionIndex]
			if action.Type != model.ActionStaticWallpaper && action.Type != model.ActionVideoWallpaper {
				continue
			}
			local := localFilePath(action.Value)
			if local == "" {
				continue
			}
			absolute, err := filepath.Abs(local)
			if err == nil {
				action.Value = absolute
			}
		}
	}
	return out
}

func (a *App) validateConfig(cfg model.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	for _, phase := range cfg.Phases {
		if strings.TrimSpace(phase.Color) != "" && !hexColorPattern.MatchString(strings.TrimSpace(phase.Color)) {
			return fmt.Errorf("phase %q color %q must be #RRGGBB", phase.Name, phase.Color)
		}
		for _, action := range phase.Actions {
			if err := a.validateAction(action); err != nil {
				return fmt.Errorf("phase %q: %w", phase.Name, err)
			}
		}
	}
	return nil
}

func (a *App) validateAction(action model.Action) error {
	if err := action.Validate(); err != nil {
		return err
	}
	switch action.Type {
	case model.ActionColorScheme:
		return validateChoice(action, a.inv.ColorSchemes, "color scheme")
	case model.ActionGlobalTheme:
		return validateChoice(action, a.inv.GlobalThemes, "global theme")
	case model.ActionAccentColor:
		if !hexColorPattern.MatchString(action.Value) {
			return fmt.Errorf("accent color %q must be #RRGGBB", action.Value)
		}
	case model.ActionPlasmaStyle:
		return validateChoice(action, a.inv.PlasmaStyles, "Plasma style")
	case model.ActionIconTheme:
		return validateChoice(action, a.inv.IconThemes, "icon theme")
	case model.ActionCursorTheme:
		return validateChoice(action, a.inv.CursorThemes, "cursor theme")
	case model.ActionWindowDecoration:
		return validateChoice(action, a.inv.WindowDecorations, "window decoration")
	case model.ActionStaticWallpaper:
		return validateExistingFile(action.Value, "static wallpaper", kde.StaticWallpaperExtensions())
	case model.ActionVideoWallpaper:
		if !a.inv.SmartVideoPlugin {
			return errors.New("video wallpaper requires Smart Video Wallpaper Reborn")
		}
		return validateExistingFile(action.Value, "video wallpaper", kde.VideoWallpaperExtensions())
	case model.ActionSDDMTheme:
		return validateChoice(action, a.inv.SDDMThemes, "SDDM theme")
	case model.ActionPlymouthTheme:
		return validateChoice(action, a.inv.PlymouthThemes, "Plymouth theme")
	}
	return nil
}

func (a *App) actionOptions() []ActionOptionView {
	add := func(action model.ActionType, choices []kde.Choice, placeholder string, warning string) ActionOptionView {
		return ActionOptionView{Type: action, Label: model.Action{Type: action}.Label(), Placeholder: placeholder, Choices: choices, Warning: warning}
	}
	options := []ActionOptionView{
		add(model.ActionAccentColor, nil, "#3daee9", ""),
		add(model.ActionStaticWallpaper, a.inv.StaticWallpapers, "/path/to/image.jpg", ""),
	}
	videoWarning := ""
	if !a.inv.SmartVideoPlugin {
		videoWarning = "Smart Video Wallpaper Reborn is not installed yet."
	}
	options = append(options, add(model.ActionVideoWallpaper, a.inv.VideoWallpapers, "/path/to/video.mp4", videoWarning))
	appendIf := func(ok bool, action model.ActionType, choices []kde.Choice) {
		if ok {
			options = append(options, add(action, choices, "installed value", ""))
		}
	}
	appendIf(len(a.inv.ColorSchemes) > 0, model.ActionColorScheme, a.inv.ColorSchemes)
	appendIf(len(a.inv.GlobalThemes) > 0, model.ActionGlobalTheme, a.inv.GlobalThemes)
	appendIf(len(a.inv.PlasmaStyles) > 0, model.ActionPlasmaStyle, a.inv.PlasmaStyles)
	appendIf(len(a.inv.IconThemes) > 0, model.ActionIconTheme, a.inv.IconThemes)
	appendIf(len(a.inv.CursorThemes) > 0, model.ActionCursorTheme, a.inv.CursorThemes)
	appendIf(len(a.inv.WindowDecorations) > 0, model.ActionWindowDecoration, a.inv.WindowDecorations)
	appendIf(len(a.inv.SDDMThemes) > 0, model.ActionSDDMTheme, a.inv.SDDMThemes)
	appendIf(len(a.inv.PlymouthThemes) > 0, model.ActionPlymouthTheme, a.inv.PlymouthThemes)
	return options
}

func validateChoice(action model.Action, choices []kde.Choice, label string) error {
	if len(choices) == 0 {
		return nil
	}
	value := strings.TrimSpace(action.Value)
	for _, choice := range choices {
		if value == choice.ID {
			return nil
		}
	}
	return fmt.Errorf("%s %q is not installed", label, value)
}

func validateExistingFile(value, label string, extensions []string) error {
	path := localFilePath(value)
	if path == "" {
		return fmt.Errorf("%s path is required", label)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s file %q is not available", label, value)
	}
	if info.IsDir() {
		return fmt.Errorf("%s path %q is a directory", label, value)
	}
	ext := strings.ToLower(filepath.Ext(path))
	if len(extensions) > 0 && !slices.Contains(extensions, ext) {
		return fmt.Errorf("%s file %q has an unsupported extension", label, value)
	}
	return nil
}

func localFilePath(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "file://") {
		uri, err := url.Parse(value)
		if err != nil {
			return ""
		}
		return uri.Path
	}
	if strings.Contains(value, "://") {
		return ""
	}
	if strings.HasPrefix(value, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return value
		}
		if value == "~" {
			return home
		}
		if strings.HasPrefix(value, "~/") {
			return filepath.Join(home, strings.TrimPrefix(value, "~/"))
		}
	}
	return value
}
