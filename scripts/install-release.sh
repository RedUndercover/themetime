#!/usr/bin/env bash
set -euo pipefail

package_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
mode="user"
with_service=false
with_privileged=false
enable_rootd=false

usage() {
  cat <<'EOF'
Usage: ./install.sh [options]

Options:
  --user              Install into ~/.local (default)
  --system            Install into /usr/local and system data directories
  --with-service      Install and start the user scheduler (user mode)
  --with-privileged   Install the restricted root helper and daemon (system mode)
  --enable-rootd      Enable and start the privileged daemon
  -h, --help          Show this help
EOF
}

while (($#)); do
  case "$1" in
    --user) mode="user" ;;
    --system) mode="system" ;;
    --with-service) with_service=true ;;
    --with-privileged) with_privileged=true ;;
    --enable-rootd) with_privileged=true; enable_rootd=true ;;
    -h|--help) usage; exit 0 ;;
    *) printf 'error: unknown option: %s\n' "$1" >&2; usage >&2; exit 2 ;;
  esac
  shift
done

if [[ "$mode" == "user" ]]; then
  if "$with_privileged"; then
    printf 'error: --with-privileged requires --system\n' >&2
    exit 2
  fi
  data_home="${XDG_DATA_HOME:-$HOME/.local/share}"
  install -Dm755 "$package_dir/bin/themetime" "$HOME/.local/bin/themetime"
  install -Dm755 "$package_dir/bin/themetime-wails" "$HOME/.local/bin/themetime-wails"
  install -Dm644 "$package_dir/share/applications/io.github.themetime.ThemeTime.desktop" \
    "$data_home/applications/io.github.themetime.ThemeTime.desktop"
  install -Dm644 "$package_dir/share/icons/hicolor/scalable/apps/io.github.themetime.ThemeTime.svg" \
    "$data_home/icons/hicolor/scalable/apps/io.github.themetime.ThemeTime.svg"
  command -v update-desktop-database >/dev/null && update-desktop-database "$data_home/applications" || true
  command -v gtk-update-icon-cache >/dev/null && gtk-update-icon-cache -f -t "$data_home/icons/hicolor" || true
  command -v kbuildsycoca6 >/dev/null && kbuildsycoca6 || true
  if "$with_service"; then
    "$HOME/.local/bin/themetime" install-user-service
  fi
  printf 'ThemeTime installed for %s. Run: %s\n' "${USER:-$(id -un)}" "$HOME/.local/bin/themetime gui"
  exit 0
fi

if ((EUID != 0)); then
  printf 'error: system installation must run as root (try sudo ./install.sh --system)\n' >&2
  exit 1
fi
if "$with_service"; then
  printf 'error: --with-service is only available with --user\n' >&2
  exit 2
fi

install -Dm755 "$package_dir/bin/themetime" /usr/local/bin/themetime
install -Dm755 "$package_dir/bin/themetime-wails" /usr/local/bin/themetime-wails
install -Dm644 "$package_dir/share/applications/io.github.themetime.ThemeTime.desktop" \
  /usr/local/share/applications/io.github.themetime.ThemeTime.desktop
install -Dm644 "$package_dir/share/icons/hicolor/scalable/apps/io.github.themetime.ThemeTime.svg" \
  /usr/local/share/icons/hicolor/scalable/apps/io.github.themetime.ThemeTime.svg

if "$with_privileged"; then
  install -Dm755 "$package_dir/libexec/themetime-rootctl" /usr/local/libexec/themetime-rootctl
  install -Dm755 "$package_dir/libexec/themetime-rootd" /usr/local/libexec/themetime-rootd
  install -Dm644 "$package_dir/share/polkit-1/actions/io.github.themetime.rootctl.policy" \
    /usr/share/polkit-1/actions/io.github.themetime.rootctl.policy
  install -Dm644 "$package_dir/lib/systemd/system/themetime-rootd.service" \
    /etc/systemd/system/themetime-rootd.service
  systemctl daemon-reload
  if "$enable_rootd"; then
    systemctl enable --now themetime-rootd.service
  fi
fi

command -v update-desktop-database >/dev/null && update-desktop-database /usr/local/share/applications || true
command -v gtk-update-icon-cache >/dev/null && gtk-update-icon-cache -f -t /usr/local/share/icons/hicolor || true
printf 'ThemeTime installed system-wide. Run: themetime gui\n'
