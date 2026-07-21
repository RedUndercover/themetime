package kde

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Inventory struct {
	Commands          map[string]bool `json:"commands"`
	ColorSchemes      []Choice        `json:"colorSchemes"`
	GlobalThemes      []Choice        `json:"globalThemes"`
	PlasmaStyles      []Choice        `json:"plasmaStyles"`
	CursorThemes      []Choice        `json:"cursorThemes"`
	IconThemes        []Choice        `json:"iconThemes"`
	WindowDecorations []Choice        `json:"windowDecorations"`
	WallpaperPlugins  []Choice        `json:"wallpaperPlugins"`
	StaticWallpapers  []Choice        `json:"staticWallpapers"`
	VideoWallpapers   []Choice        `json:"videoWallpapers"`
	SDDMThemes        []Choice        `json:"sddmThemes"`
	PlymouthThemes    []Choice        `json:"plymouthThemes"`
	SmartVideoPlugin  bool            `json:"smartVideoPlugin"`
}

type Choice struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Current bool   `json:"current,omitempty"`
}

func Discover(ctx context.Context, r Runner) Inventory {
	commands := []string{
		"plasma-apply-colorscheme",
		"plasma-apply-cursortheme",
		"plasma-apply-desktoptheme",
		"plasma-apply-lookandfeel",
		"plasma-apply-wallpaperimage",
		"kwriteconfig6",
		"kreadconfig6",
		"qdbus6",
		"kpackagetool6",
		"kbuildsycoca6",
		"plymouth-set-default-theme",
		"systemctl",
		"pkexec",
	}
	inv := Inventory{Commands: map[string]bool{}}
	for _, command := range commands {
		inv.Commands[command] = commandExists(r, command)
	}
	inv.ColorSchemes = parseStarChoices(runIgnoreError(ctx, r, "plasma-apply-colorscheme", "--list-schemes"))
	inv.GlobalThemes = parseLineChoices(runIgnoreError(ctx, r, "plasma-apply-lookandfeel", "--list"))
	inv.PlasmaStyles = parseStarChoices(runIgnoreError(ctx, r, "plasma-apply-desktoptheme", "--list-themes"))
	inv.CursorThemes = parseCursorChoices(runIgnoreError(ctx, r, "plasma-apply-cursortheme", "--list-themes"))
	inv.IconThemes = discoverIconThemes()
	inv.WindowDecorations = discoverKPackage(ctx, r, "KWin/Decoration")
	inv.WallpaperPlugins = discoverWallpapers(ctx, r)
	inv.StaticWallpapers = discoverMediaFiles(staticWallpaperRoots(), []string{".avif", ".bmp", ".jpeg", ".jpg", ".png", ".svg", ".webp"}, 250)
	inv.VideoWallpapers = discoverMediaFiles(videoWallpaperRoots(), []string{".avi", ".m4v", ".mkv", ".mov", ".mp4", ".webm"}, 250)
	inv.SDDMThemes = discoverSDDMThemes()
	inv.PlymouthThemes = parseLineChoices(runIgnoreError(ctx, r, "plymouth-set-default-theme", "--list"))
	for _, choice := range inv.WallpaperPlugins {
		if choice.ID == SmartVideoWallpaperPlugin {
			inv.SmartVideoPlugin = true
			break
		}
	}
	return inv
}

func runIgnoreError(ctx context.Context, r Runner, name string, args ...string) string {
	if !commandExists(r, name) {
		return ""
	}
	out, _ := r.Run(ctx, name, args...)
	return out
}

func parseStarChoices(out string) []Choice {
	var choices []Choice
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "* ") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "* "))
		current := strings.Contains(line, "(current") || strings.Contains(line, "(Current")
		line = strings.TrimSpace(strings.Split(line, "(")[0])
		if line != "" {
			choices = append(choices, Choice{ID: line, Name: line, Current: current})
		}
	}
	return choices
}

func parseLineChoices(out string) []Choice {
	var choices []Choice
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "Usage:") || strings.Contains(line, "available") {
			continue
		}
		choices = append(choices, Choice{ID: line, Name: line})
	}
	return choices
}

func parseCursorChoices(out string) []Choice {
	var choices []Choice
	scanner := bufio.NewScanner(strings.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "* ") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "* "))
		current := strings.Contains(line, "(Current")
		line = strings.TrimSpace(strings.Split(line, "(")[0])
		id := line
		if start := strings.LastIndex(line, "["); start >= 0 && strings.HasSuffix(line, "]") {
			id = strings.TrimSuffix(line[start+1:], "]")
			line = strings.TrimSpace(line[:start])
		}
		if id != "" {
			choices = append(choices, Choice{ID: id, Name: line, Current: current})
		}
	}
	return choices
}

func discoverIconThemes() []Choice {
	var roots []string
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots, filepath.Join(home, ".local", "share", "icons"))
	}
	roots = append(roots, "/usr/share/icons")
	var choices []Choice
	seen := map[string]bool{}
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			id := entry.Name()
			if seen[id] {
				continue
			}
			index := filepath.Join(root, id, "index.theme")
			if _, err := os.Stat(index); err != nil {
				continue
			}
			seen[id] = true
			choices = append(choices, Choice{ID: id, Name: readThemeName(index, id)})
		}
	}
	sort.Slice(choices, func(i, j int) bool {
		return strings.ToLower(choices[i].Name) < strings.ToLower(choices[j].Name)
	})
	return choices
}

func readThemeName(path, fallback string) string {
	file, err := os.Open(path)
	if err != nil {
		return fallback
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Name=") {
			name := strings.TrimSpace(strings.TrimPrefix(line, "Name="))
			if name != "" {
				return name
			}
		}
	}
	return fallback
}

func discoverKPackage(ctx context.Context, r Runner, packageType string) []Choice {
	var choices []Choice
	for _, global := range []bool{true, false} {
		args := []string{"--type", packageType, "--list"}
		if global {
			args = append([]string{"--global"}, args...)
		}
		out := runIgnoreError(ctx, r, "kpackagetool6", args...)
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "Listing ") {
				continue
			}
			choices = appendUniqueChoice(choices, Choice{ID: line, Name: line})
		}
	}
	return choices
}

func discoverWallpapers(ctx context.Context, r Runner) []Choice {
	choices := discoverKPackage(ctx, r, "Plasma/Wallpaper")
	for _, root := range []string{"/usr/share/plasma/wallpapers"} {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				choices = appendUniqueChoice(choices, Choice{ID: entry.Name(), Name: entry.Name()})
			}
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		root := filepath.Join(home, ".local", "share", "plasma", "wallpapers")
		entries, err := os.ReadDir(root)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					choices = appendUniqueChoice(choices, Choice{ID: entry.Name(), Name: entry.Name()})
				}
			}
		}
	}
	return choices
}

func discoverSDDMThemes() []Choice {
	var choices []Choice
	for _, root := range []string{"/usr/share/sddm/themes", "/usr/local/share/sddm/themes"} {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				choices = appendUniqueChoice(choices, Choice{ID: entry.Name(), Name: entry.Name()})
			}
		}
	}
	return choices
}

func discoverMediaFiles(roots []string, extensions []string, limit int) []Choice {
	exts := map[string]bool{}
	for _, ext := range extensions {
		exts[strings.ToLower(ext)] = true
	}
	var choices []Choice
	seen := map[string]bool{}
	for _, root := range roots {
		if len(choices) >= limit {
			break
		}
		info, err := os.Stat(root)
		if err != nil || !info.IsDir() {
			continue
		}
		_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if len(choices) >= limit {
				return filepath.SkipAll
			}
			if entry.IsDir() {
				if mediaSearchDepth(root, path) > 3 {
					return filepath.SkipDir
				}
				if path != root && strings.HasPrefix(entry.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if !exts[strings.ToLower(filepath.Ext(path))] || seen[path] {
				return nil
			}
			seen[path] = true
			name := filepath.Base(path)
			choices = append(choices, Choice{ID: path, Name: name})
			return nil
		})
	}
	sort.Slice(choices, func(i, j int) bool {
		return strings.ToLower(choices[i].Name) < strings.ToLower(choices[j].Name)
	})
	return choices
}

func staticWallpaperRoots() []string {
	roots := []string{"/usr/share/wallpapers", "/usr/share/backgrounds"}
	if home, err := os.UserHomeDir(); err == nil {
		roots = append([]string{
			filepath.Join(home, "Pictures"),
			filepath.Join(home, "Wallpapers"),
		}, roots...)
	}
	return roots
}

func videoWallpaperRoots() []string {
	var roots []string
	if home, err := os.UserHomeDir(); err == nil {
		roots = append(roots, filepath.Join(home, "Videos"), filepath.Join(home, "Wallpapers"))
	}
	return roots
}

func mediaSearchDepth(root, path string) int {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." {
		return 0
	}
	return strings.Count(rel, string(filepath.Separator)) + 1
}

func appendUniqueChoice(choices []Choice, next Choice) []Choice {
	for _, choice := range choices {
		if choice.ID == next.ID {
			return choices
		}
	}
	return append(choices, next)
}
