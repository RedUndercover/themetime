package kde

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/themetime/themetime/internal/config"
	"github.com/themetime/themetime/internal/model"
)

var ErrPrivilegedAction = errors.New("action requires privileged helper")

type Applier struct {
	Runner      Runner
	SnapshotDir string
}

type Result struct {
	Action  model.Action `json:"action"`
	Skipped bool         `json:"skipped"`
	Message string       `json:"message"`
	Error   string       `json:"error,omitempty"`
}

func NewApplier(snapshotDir string) Applier {
	return Applier{
		Runner:      ExecRunner{},
		SnapshotDir: snapshotDir,
	}
}

func (a Applier) ApplyPhase(ctx context.Context, phase model.Phase) []Result {
	results := make([]Result, 0, len(phase.Actions))
	actions := orderedActions(phase.Actions)
	for _, action := range actions {
		result := Result{Action: action}
		if action.IsPrivileged() {
			result.Skipped = true
			result.Message = ErrPrivilegedAction.Error()
			results = append(results, result)
			continue
		}
		if err := a.ApplyAction(ctx, action); err != nil {
			result.Error = err.Error()
		} else {
			result.Message = "applied"
		}
		results = append(results, result)
	}
	return results
}

func (a Applier) ApplyAction(ctx context.Context, action model.Action) error {
	r := a.Runner
	if r == nil {
		r = ExecRunner{}
	}
	switch action.Type {
	case model.ActionGlobalTheme:
		return runRequired(ctx, r, "plasma-apply-lookandfeel", "--apply", action.Value)
	case model.ActionColorScheme:
		args := []string{}
		if accent := strings.TrimSpace(action.Values["accent"]); accent != "" {
			args = append(args, "--accent-color", accent)
		}
		args = append(args, action.Value)
		return runRequired(ctx, r, "plasma-apply-colorscheme", args...)
	case model.ActionAccentColor:
		scheme := action.Values["colorScheme"]
		if scheme == "" {
			scheme = readConfig(ctx, r, "kdeglobals", "General", "ColorScheme")
		}
		if scheme == "" {
			return fmt.Errorf("cannot apply accent color without current color scheme")
		}
		return runRequired(ctx, r, "plasma-apply-colorscheme", "--accent-color", action.Value, scheme)
	case model.ActionPlasmaStyle:
		return runRequired(ctx, r, "plasma-apply-desktoptheme", action.Value)
	case model.ActionCursorTheme:
		args := []string{}
		if size := strings.TrimSpace(action.Values["size"]); size != "" {
			args = append(args, "--size", size)
		}
		args = append(args, action.Value)
		return runRequired(ctx, r, "plasma-apply-cursortheme", args...)
	case model.ActionIconTheme:
		if err := a.snapshot("kdeglobals"); err != nil {
			return err
		}
		if err := runRequired(ctx, r, "kwriteconfig6", "--file", "kdeglobals", "--group", "Icons", "--key", "Theme", action.Value); err != nil {
			return err
		}
		runOptional(ctx, r, "kbuildsycoca6", "--noincremental")
		return nil
	case model.ActionWindowDecoration:
		if err := a.snapshot("kwinrc"); err != nil {
			return err
		}
		library := valueOrDefault(action.Values["library"], "org.kde.breeze")
		theme := valueOrDefault(action.Value, action.Values["theme"])
		if err := runRequired(ctx, r, "kwriteconfig6", "--file", "kwinrc", "--group", "org.kde.kdecoration2", "--key", "library", library); err != nil {
			return err
		}
		if err := runRequired(ctx, r, "kwriteconfig6", "--file", "kwinrc", "--group", "org.kde.kdecoration2", "--key", "theme", theme); err != nil {
			return err
		}
		runOptional(ctx, r, "qdbus6", "org.kde.KWin", "/KWin", "reconfigure")
		return nil
	case model.ActionFontProfile:
		return a.applyFonts(ctx, r, action.Values)
	case model.ActionStaticWallpaper:
		wallpaperPath := expandUserPath(action.Value)
		values := map[string]string{
			"Image":    fileURI(wallpaperPath),
			"FillMode": valueOrDefault(action.Values["fillMode"], "2"),
		}
		if action.Screen == "" {
			if commandExists(r, "plasma-apply-wallpaperimage") {
				args := []string{}
				if fill := action.Values["fillMode"]; fill != "" {
					args = append(args, "--fill-mode", fill)
				}
				args = append(args, wallpaperPath)
				return runRequired(ctx, r, "plasma-apply-wallpaperimage", args...)
			}
		}
		return applyWallpaperScript(ctx, r, "org.kde.image", action.Screen, values)
	case model.ActionVideoWallpaper:
		values, err := videoWallpaperValues(expandUserPath(action.Value), action.Values)
		if err != nil {
			return err
		}
		return applyWallpaperScript(ctx, r, SmartVideoWallpaperPlugin, action.Screen, values)
	case model.ActionCustomCommand:
		return runRequired(ctx, r, "/bin/sh", "-c", action.Value)
	case model.ActionSDDMTheme, model.ActionPlymouthTheme:
		return ErrPrivilegedAction
	default:
		return fmt.Errorf("unsupported action %q", action.Type)
	}
}

func (a Applier) applyFonts(ctx context.Context, r Runner, values map[string]string) error {
	if err := a.snapshot("kdeglobals"); err != nil {
		return err
	}
	keys := map[string]string{
		"font":            "font",
		"fixed":           "fixed",
		"smallestFont":    "smallestReadableFont",
		"toolBarFont":     "toolBarFont",
		"menuFont":        "menuFont",
		"activeFont":      "activeFont",
		"desktopFont":     "desktopFont",
		"taskbarFont":     "taskbarFont",
		"windowTitle":     "activeFont",
		"windowTitleFont": "activeFont",
	}
	for input, key := range keys {
		value := strings.TrimSpace(values[input])
		if value == "" {
			continue
		}
		if err := runRequired(ctx, r, "kwriteconfig6", "--file", "kdeglobals", "--group", "General", "--key", key, value); err != nil {
			return err
		}
	}
	runOptional(ctx, r, "qdbus6", "org.kde.KWin", "/KWin", "reconfigure")
	return nil
}

func (a Applier) snapshot(file string) error {
	if a.SnapshotDir == "" {
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	path := filepath.Join(home, ".config", file)
	_, err = config.SnapshotFile(a.SnapshotDir, path)
	return err
}

func orderedActions(actions []model.Action) []model.Action {
	out := append([]model.Action(nil), actions...)
	sort.SliceStable(out, func(i, j int) bool {
		return model.ActionPriority(out[i].Type) < model.ActionPriority(out[j].Type)
	})
	return out
}

func runRequired(ctx context.Context, r Runner, name string, args ...string) error {
	if _, err := r.LookPath(name); err != nil && !strings.Contains(name, "/") {
		return fmt.Errorf("%s is not installed", name)
	}
	_, err := r.Run(ctx, name, args...)
	return err
}

func runOptional(ctx context.Context, r Runner, name string, args ...string) {
	if _, err := r.LookPath(name); err == nil || strings.Contains(name, "/") {
		_, _ = r.Run(ctx, name, args...)
	}
}

func readConfig(ctx context.Context, r Runner, file, group, key string) string {
	out, err := r.Run(ctx, "kreadconfig6", "--file", file, "--group", group, "--key", key)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func fileURI(path string) string {
	if strings.HasPrefix(path, "file://") || strings.Contains(path, "://") {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	u := url.URL{Scheme: "file", Path: abs}
	return u.String()
}

func expandUserPath(value string) string {
	value = strings.TrimSpace(value)
	if value != "~" && !strings.HasPrefix(value, "~/") {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return value
	}
	if value == "~" {
		return home
	}
	return filepath.Join(home, strings.TrimPrefix(value, "~/"))
}
