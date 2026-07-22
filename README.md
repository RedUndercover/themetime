# ThemeTime

ThemeTime is a KDE Plasma 6 scheduler for desktop themes and wallpapers. It
can start phases at fixed times or relative to sunrise, sunset, twilight, and
solar noon. Its Solar Observatory interface shows the day on a 24-hour
timeline and lets you preview changes before enabling automatic scheduling.

ThemeTime can change KDE themes, colors, accents, Plasma styles, icons,
cursors, decorations, fonts, static or video wallpapers, and run user-defined
commands. An optional restricted service handles SDDM and Plymouth themes.
Normal desktop scheduling does not require root access.

[![CI](https://github.com/RedUndercover/themetime/actions/workflows/ci.yml/badge.svg)](https://github.com/RedUndercover/themetime/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Install a release

ThemeTime currently targets Linux x86-64 with KDE Plasma 6. Tagged releases
provide an Arch Linux package and a portable archive.

On Arch Linux:

```sh
sudo pacman -U themetime_VERSION_linux_amd64.pkg.tar.zst
systemctl --user enable --now themetime.service
themetime doctor
```

For the portable archive:

```sh
sha256sum --ignore-missing --check checksums.txt
mkdir themetime-VERSION
tar -xzf themetime-VERSION-linux-x86_64.tar.gz -C themetime-VERSION
cd themetime-VERSION
./themetime doctor
./themetime gui
```

Replace `VERSION` with the release number. The portable archive runs in place
and does not install desktop integration or a systemd service. See
[Getting started](docs/getting-started.md) for dependencies, source builds,
installation, and removal.

## Build from source

The full application requires Go 1.25+, Node.js/npm, a C compiler,
`pkg-config`, GTK 3, WebKitGTK, and KDE Plasma command-line tools.

```sh
make build
./bin/themetime doctor
./bin/themetime gui
```

Install the user binaries and desktop integration with:

```sh
make install-user-assets
./bin/themetime install-user-service
```

## First schedule

1. Open **Location** and enter coordinates plus an IANA timezone.
2. Select a phase and anchor it to a solar event or fixed time.
3. Add the desired appearance actions.
4. Save, then use **Apply now** to test the phase.
5. Install the user service when the schedule is ready.

The default configuration demonstrates light and dark phases around sunrise
and sunset. Actions are layered by setting, so a wallpaper-only phase does not
discard the current colors, icons, or other appearance choices.

## Documentation

- [Documentation index](docs/README.md)
- [Getting started](docs/getting-started.md)
- [How-to guide and recipes](docs/how-to.md)
- [User guide](docs/user-guide.md)
- [Configuration reference](docs/configuration.md)
- [Action reference](docs/actions.md)
- [CLI reference](docs/cli.md)
- [Services and privileged actions](docs/services-and-privileged-actions.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Architecture](docs/architecture.md)
- [Development guide](docs/development.md)
- [Packaging and releases](docs/packaging-and-releases.md)

Common commands:

```sh
themetime show-config
themetime doctor
themetime daemon --once
themetime apply --phase evening
```

ThemeTime stores user configuration under `~/.config/themetime/` and state
under `~/.local/state/themetime/` by default. XDG environment variables take
precedence.
