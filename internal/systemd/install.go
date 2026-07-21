package systemd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const userUnitName = "themetime.service"

func InstallUserService(binary string, enableNow bool) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	binary, err = resolveDaemonBinary(binary)
	if err != nil {
		return "", err
	}
	unitDir := filepath.Join(home, ".config", "systemd", "user")
	if err := os.MkdirAll(unitDir, 0o755); err != nil {
		return "", err
	}
	unitPath := filepath.Join(unitDir, userUnitName)
	content := fmt.Sprintf(`[Unit]
Description=ThemeTime KDE schedule daemon
After=graphical-session.target plasma-plasmashell.service

[Service]
Type=simple
ExecStart=%s daemon
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=default.target
`, systemdQuotePath(binary))
	if err := os.WriteFile(unitPath, []byte(content), 0o644); err != nil {
		return "", err
	}
	if out, err := exec.Command("systemctl", "--user", "daemon-reload").CombinedOutput(); err != nil {
		return "", fmt.Errorf("systemctl --user daemon-reload: %w: %s", err, strings.TrimSpace(string(out)))
	}
	if enableNow {
		if out, err := exec.Command("systemctl", "--user", "enable", "--now", userUnitName).CombinedOutput(); err != nil {
			return "", fmt.Errorf("systemctl --user enable --now %s: %w: %s", userUnitName, err, strings.TrimSpace(string(out)))
		}
	}
	return unitPath, nil
}

func resolveDaemonBinary(binary string) (string, error) {
	if strings.TrimSpace(binary) != "" {
		return normalizeBinaryPath(binary, true)
	}
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return resolveDaemonBinaryFromExecutable("", executable)
}

func resolveDaemonBinaryFromExecutable(binary, executable string) (string, error) {
	if strings.TrimSpace(binary) != "" {
		return normalizeBinaryPath(binary, true)
	}
	if executable == "" {
		return "", errors.New("cannot infer daemon binary path")
	}
	dir := filepath.Dir(executable)
	for _, name := range []string{"themetime", "themetime.exe"} {
		candidate := filepath.Join(dir, name)
		if path, err := normalizeBinaryPath(candidate, false); err == nil {
			return path, nil
		}
	}
	base := strings.ToLower(filepath.Base(executable))
	if base == "themetime-wails" || base == "themetime-wails.exe" {
		if path, err := exec.LookPath("themetime"); err == nil {
			return normalizeBinaryPath(path, false)
		}
		return "", fmt.Errorf("cannot install user service from %s because the themetime daemon binary was not found next to it", executable)
	}
	if path, err := normalizeBinaryPath(executable, false); err == nil {
		return path, nil
	}
	if path, err := exec.LookPath("themetime"); err == nil {
		return normalizeBinaryPath(path, false)
	}
	return "", errors.New("cannot infer daemon binary path; build ThemeTime or pass --binary")
}

func normalizeBinaryPath(binary string, allowTransient bool) (string, error) {
	if !strings.ContainsRune(binary, os.PathSeparator) {
		path, err := exec.LookPath(binary)
		if err != nil {
			return "", err
		}
		binary = path
	}
	abs, err := filepath.Abs(binary)
	if err != nil {
		return "", err
	}
	if !allowTransient && isTransientExecutable(abs) {
		return "", fmt.Errorf("%s looks like a temporary go run executable", abs)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory", abs)
	}
	if info.Mode().Perm()&0o111 == 0 {
		return "", fmt.Errorf("%s is not executable", abs)
	}
	return abs, nil
}

func isTransientExecutable(path string) bool {
	return strings.Contains(path, string(filepath.Separator)+"go-build")
}

func systemdQuotePath(path string) string {
	return strconv.Quote(strings.ReplaceAll(path, "%", "%%"))
}
