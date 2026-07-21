#!/usr/bin/env bash
set -euo pipefail

mode="user"
purge_config=false

usage() {
  cat <<'EOF'
Usage: ./uninstall.sh [--user|--system] [--purge-config]

Configuration and state are preserved unless --purge-config is supplied.
System mode also removes the optional privileged helper and daemon.
EOF
}

while (($#)); do
  case "$1" in
    --user) mode="user" ;;
    --system) mode="system" ;;
    --purge-config) purge_config=true ;;
    -h|--help) usage; exit 0 ;;
    *) printf 'error: unknown option: %s\n' "$1" >&2; usage >&2; exit 2 ;;
  esac
  shift
done

if [[ "$mode" == "user" ]]; then
  data_home="${XDG_DATA_HOME:-$HOME/.local/share}"
  config_home="${XDG_CONFIG_HOME:-$HOME/.config}"
  state_home="${XDG_STATE_HOME:-$HOME/.local/state}"
  systemctl --user disable --now themetime.service >/dev/null 2>&1 || true
  rm -f -- \
    "$HOME/.local/bin/themetime" \
    "$HOME/.local/bin/themetime-wails" \
    "$data_home/applications/io.github.themetime.ThemeTime.desktop" \
    "$data_home/icons/hicolor/scalable/apps/io.github.themetime.ThemeTime.svg" \
    "$config_home/systemd/user/themetime.service"
  systemctl --user daemon-reload >/dev/null 2>&1 || true
  if "$purge_config"; then
    rm -rf -- "$config_home/themetime" "$state_home/themetime"
  fi
  command -v update-desktop-database >/dev/null && update-desktop-database "$data_home/applications" || true
  command -v gtk-update-icon-cache >/dev/null && gtk-update-icon-cache -f -t "$data_home/icons/hicolor" || true
  command -v kbuildsycoca6 >/dev/null && kbuildsycoca6 || true
  printf 'ThemeTime user installation removed.\n'
  exit 0
fi

if ((EUID != 0)); then
  printf 'error: system uninstall must run as root (try sudo ./uninstall.sh --system)\n' >&2
  exit 1
fi

systemctl disable --now themetime-rootd.service >/dev/null 2>&1 || true
rm -f -- \
  /usr/local/bin/themetime \
  /usr/local/bin/themetime-wails \
  /usr/local/libexec/themetime-rootctl \
  /usr/local/libexec/themetime-rootd \
  /usr/local/share/applications/io.github.themetime.ThemeTime.desktop \
  /usr/local/share/icons/hicolor/scalable/apps/io.github.themetime.ThemeTime.svg \
  /usr/share/polkit-1/actions/io.github.themetime.rootctl.policy \
  /etc/systemd/system/themetime-rootd.service
systemctl daemon-reload
if "$purge_config"; then
  rm -rf -- /etc/themetime /var/lib/themetime
fi
command -v update-desktop-database >/dev/null && update-desktop-database /usr/local/share/applications || true
command -v gtk-update-icon-cache >/dev/null && gtk-update-icon-cache -f -t /usr/local/share/icons/hicolor || true
printf 'ThemeTime system installation removed.\n'
