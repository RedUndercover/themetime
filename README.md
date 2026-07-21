# ThemeTime

ThemeTime is a KDE Plasma 6 scheduler that changes the look of your desktop at
fixed times or relative to sunrise, sunset, twilight, and solar noon. Its Solar
Observatory interface shows the whole day on a 24-hour sky timeline, including
phase bands, solar markers, and the currently active and next phases.

ThemeTime can apply:

- global themes, color schemes, accent colors, Plasma styles, icons, cursors,
  window decorations, and font profiles;
- static wallpapers and videos through Smart Video Wallpaper Reborn;
- user-defined shell commands;
- SDDM and Plymouth themes through a deliberately restricted privileged
  service.

Rules are layered rather than mutually exclusive. A video-only rule changes the
wallpaper while retaining the most recently scheduled colors, icons, cursors,
and other appearance settings.

The main scheduler runs entirely in the user session. Root access is optional
and only used for the two explicitly privileged action types.

[![CI](https://github.com/themetime/themetime/actions/workflows/ci.yml/badge.svg)](https://github.com/themetime/themetime/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Install a release

Tagged releases contain native Linux x86-64 binaries, desktop integration,
installers, documentation, and a SHA-256 checksum:

```sh
sha256sum --check themetime-0.1.0-linux-x86_64.tar.gz.sha256
tar -xzf themetime-0.1.0-linux-x86_64.tar.gz
cd themetime-0.1.0-linux-x86_64
./install.sh --user --with-service
themetime doctor
```

Release binaries require a compatible KDE Plasma 6, GTK 3, WebKitGTK 4.1, and
Ayatana App Indicator runtime. See
[Packaging and releases](docs/packaging-and-releases.md) for system installs,
privileged components, uninstallation, checksums, and publishing tags.

## Documentation

- [Documentation index](docs/README.md)
- [Getting started](docs/getting-started.md)
- [User guide](docs/user-guide.md)
- [Configuration reference](docs/configuration.md)
- [Action reference](docs/actions.md)
- [CLI reference](docs/cli.md)
- [Services and privileged actions](docs/services-and-privileged-actions.md)
- [Architecture](docs/architecture.md)
- [Development guide](docs/development.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Packaging and releases](docs/packaging-and-releases.md)
- [Task-oriented wiki](wiki/Home.md)

## Requirements

ThemeTime targets Linux with KDE Plasma 6. Building it requires:

- Go 1.25 or newer;
- Node.js and npm;
- a C compiler and `pkg-config`;
- WebKitGTK 4.1 or 4.0 development files for the Wails interface;
- KDE command-line tools such as `kwriteconfig6`, `kreadconfig6`, `qdbus6`,
  and `plasma-apply-colorscheme`.

Optional KDE utilities unlock additional action types. Run `themetime doctor`
after building to see exactly what is present on the current system.

## Build and run

```sh
make build
./bin/themetime doctor
./bin/themetime gui
```

`make build` compiles four executables in `bin/`:

| Executable | Purpose |
| --- | --- |
| `themetime` | User CLI and scheduler daemon |
| `themetime-wails` | Solar Observatory desktop interface and tray |
| `themetime-rootctl` | Polkit-authorized schedule installer |
| `themetime-rootd` | Restricted system scheduler for SDDM/Plymouth |

The GUI can also be started directly with `./bin/themetime-wails`. Closing its
window hides it to the KDE system tray; choose **Quit** in the tray menu to stop
it completely.

For a source-tree development run:

```sh
go run ./cmd/themetime gui
```

## First schedule

1. Start the GUI and open **Location**.
2. Enter a label, latitude, longitude, and IANA timezone such as
   `America/New_York`, then preview and save the location.
3. Add or select a phase on the timeline.
4. Click a solar marker, or choose **Fixed time**, and adjust its offset.
5. Add the desired actions and save the schedule.
6. Install the background scheduler with:

   ```sh
   ./bin/themetime install-user-service
   ```

The default configuration contains a light morning phase at sunrise +20
minutes and a dark evening phase at sunset -30 minutes.

## Configuration and data

ThemeTime follows XDG paths. Defaults are shown when the corresponding XDG
variable is unset.

| Data | Default path |
| --- | --- |
| User configuration | `~/.config/themetime/config.json` |
| Scheduler state | `~/.local/state/themetime/state.json` |
| Configuration snapshots | `~/.local/state/themetime/snapshots/` |
| User systemd unit | `~/.config/systemd/user/themetime.service` |
| Privileged schedule | `/etc/themetime/privileged-schedule.json` |
| Privileged state | `/var/lib/themetime/state.json` |

Set `XDG_CONFIG_HOME` or `XDG_STATE_HOME` to relocate the user files. See the
[configuration reference](docs/configuration.md) before editing JSON by hand.

## Background service

Install and immediately start the per-user scheduler:

```sh
./bin/themetime install-user-service
systemctl --user status themetime.service
```

The installer records the absolute path to the current `themetime` executable.
If you move the binary later, run the installer again. To install the unit
without starting it, use `--now=false`.

## Video wallpapers

Video actions use the Plasma 6 plugin:

```text
luisbocanegra.smart.video.wallpaper.reborn
```

Install Smart Video Wallpaper Reborn before applying video actions. ThemeTime
keeps a video action in the schedule if the plugin is absent, and
`themetime doctor` reports the missing dependency.

## Privileged actions

The privileged path only accepts `sddmTheme` and `plymouthTheme`. It rejects
custom commands and every user-session action.

```sh
make build
sudo make install-root-assets
sudo systemctl enable --now themetime-rootd.service
./bin/themetime install-privileged-schedule
```

The Polkit policy authorizes `/usr/local/libexec/themetime-rootctl`; keep the
helper at that location unless the policy is updated as well. Read
[services and privileged actions](docs/services-and-privileged-actions.md) for
the security model and operational details.

## Common commands

```sh
./bin/themetime show-config
./bin/themetime doctor
./bin/themetime daemon --once
./bin/themetime apply --phase evening
make docs-check
make test
make release-check
```

Some KDE changes, especially icons and fonts, only fully affect newly launched
applications or the next login session.
