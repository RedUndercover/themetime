package doctor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/themetime/themetime/internal/config"
	"github.com/themetime/themetime/internal/kde"
)

type Check struct {
	Name    string `json:"name"`
	OK      bool   `json:"ok"`
	Detail  string `json:"detail"`
	Warning bool   `json:"warning,omitempty"`
}

func Run(ctx context.Context) []Check {
	r := kde.ExecRunner{}
	inv := kde.Discover(ctx, r)
	var checks []Check
	required := []string{"plasma-apply-colorscheme", "kwriteconfig6", "kreadconfig6", "qdbus6"}
	for _, command := range required {
		checks = append(checks, Check{
			Name:   command,
			OK:     inv.Commands[command],
			Detail: detailCommand(command, inv.Commands[command]),
		})
	}
	optional := []string{"plasma-apply-lookandfeel", "plasma-apply-desktoptheme", "plasma-apply-cursortheme", "plasma-apply-wallpaperimage", "kpackagetool6", "pkexec"}
	for _, command := range optional {
		checks = append(checks, Check{
			Name:    command,
			OK:      inv.Commands[command],
			Warning: !inv.Commands[command],
			Detail:  detailCommand(command, inv.Commands[command]),
		})
	}
	checks = append(checks, Check{
		Name:    "Smart Video Wallpaper Reborn",
		OK:      inv.SmartVideoPlugin,
		Warning: !inv.SmartVideoPlugin,
		Detail:  smartVideoDetail(inv.SmartVideoPlugin),
	})
	checks = append(checks, wailsWebKitCheck(ctx))
	paths, err := config.UserPaths()
	if err == nil {
		checks = append(checks, Check{Name: "Config path", OK: true, Detail: paths.Config})
	} else {
		checks = append(checks, Check{Name: "Config path", OK: false, Detail: err.Error()})
	}
	checks = append(checks, Check{Name: "KDE session", OK: isKDE(), Warning: !isKDE(), Detail: kdeSessionDetail()})
	checks = append(checks, Check{Name: "User service", OK: userServiceExists(), Warning: !userServiceExists(), Detail: userServiceDetail()})
	checks = append(checks, Check{Name: "Root helper", OK: rootHelperExists(), Warning: !rootHelperExists(), Detail: rootHelperDetail()})
	return checks
}

func Format(checks []Check) string {
	var b strings.Builder
	for _, check := range checks {
		marker := "OK"
		if !check.OK && check.Warning {
			marker = "WARN"
		} else if !check.OK {
			marker = "FAIL"
		}
		fmt.Fprintf(&b, "[%s] %s: %s\n", marker, check.Name, check.Detail)
	}
	return b.String()
}

func detailCommand(command string, ok bool) string {
	if !ok {
		return "not found in PATH"
	}
	path, err := exec.LookPath(command)
	if err != nil {
		return "found"
	}
	return path
}

func smartVideoDetail(ok bool) string {
	if ok {
		return "wallpaper plugin is installed"
	}
	return "install Smart Video Wallpaper Reborn for scheduled video wallpapers"
}

func wailsWebKitCheck(ctx context.Context) Check {
	if _, err := exec.LookPath("pkg-config"); err != nil {
		return Check{
			Name:    "Wails WebKitGTK",
			OK:      false,
			Warning: true,
			Detail:  "pkg-config not found; install pkgconf and WebKitGTK development files",
		}
	}
	for _, pkg := range []string{"webkit2gtk-4.1", "webkit2gtk-4.0"} {
		out, err := exec.CommandContext(ctx, "pkg-config", "--modversion", pkg).Output()
		if err == nil {
			return Check{
				Name:   "Wails WebKitGTK",
				OK:     true,
				Detail: fmt.Sprintf("%s %s", pkg, strings.TrimSpace(string(out))),
			}
		}
	}
	return Check{
		Name:    "Wails WebKitGTK",
		OK:      false,
		Warning: true,
		Detail:  "install webkit2gtk on Arch/CachyOS, libwebkit2gtk-4.0-dev on Debian/Ubuntu, or the equivalent WebKitGTK development package",
	}
}

func isKDE() bool {
	return strings.Contains(os.Getenv("XDG_CURRENT_DESKTOP"), "KDE") || os.Getenv("KDE_SESSION_VERSION") != ""
}

func kdeSessionDetail() string {
	return fmt.Sprintf("XDG_CURRENT_DESKTOP=%s KDE_SESSION_VERSION=%s", os.Getenv("XDG_CURRENT_DESKTOP"), os.Getenv("KDE_SESSION_VERSION"))
}

func userServiceExists() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(home, ".config", "systemd", "user", "themetime.service"))
	return err == nil
}

func userServiceDetail() string {
	if userServiceExists() {
		return "installed"
	}
	return "run `themetime install-user-service`"
}

func rootHelperExists() bool {
	info, err := os.Stat("/usr/local/libexec/themetime-rootctl")
	return err == nil && !info.IsDir() && info.Mode().Perm()&0o111 != 0
}

func rootHelperDetail() string {
	if rootHelperExists() {
		return "installed at /usr/local/libexec/themetime-rootctl"
	}
	return "not installed at /usr/local/libexec/themetime-rootctl; run `sudo make install-root-assets`"
}
