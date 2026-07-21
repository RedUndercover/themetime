# Installation

## Install a GitHub release

For a tagged release, download the archive and checksum, then run:

```sh
sha256sum --check themetime-0.1.0-linux-x86_64.tar.gz.sha256
tar -xzf themetime-0.1.0-linux-x86_64.tar.gz
cd themetime-0.1.0-linux-x86_64
./install.sh --user --with-service
```

This installs the CLI, Solar Observatory, launcher, icon, and user scheduler.
The optional root components are documented in
[Privileged themes](Privileged-Themes).

## 1. Check the platform

ThemeTime targets Linux with KDE Plasma 6. You need Go 1.25+, Node.js/npm, a C
compiler, `pkg-config`, WebKitGTK development files, and Plasma command-line
tools.

Package names vary, so install the normal development toolchain for your
distribution and let ThemeTime identify any remaining gaps later.

## 2. Build

From the repository root:

```sh
make build
```

Successful output appears in `bin/`:

```text
themetime
themetime-wails
themetime-rootctl
themetime-rootd
```

Run diagnostics:

```sh
./bin/themetime doctor
```

Required KDE tools should report as present. Optional tool warnings only matter
for the action types that use them.

## 3. Start ThemeTime

```sh
./bin/themetime gui
```

The CLI finds the Wails app beside it. The first run creates the default config
and opens the Solar Observatory.

If the GUI launcher cannot find its companion:

```sh
THEMETIME_GUI="$PWD/bin/themetime-wails" ./bin/themetime gui
```

## 4. Optionally install for the current user

Install the binaries, KDE launcher, and solar icon together:

```sh
make install-user-assets
```

Or install each file manually:

```sh
install -Dm755 bin/themetime ~/.local/bin/themetime
install -Dm755 bin/themetime-wails ~/.local/bin/themetime-wails
install -Dm644 assets/desktop/io.github.themetime.ThemeTime.desktop \
  ~/.local/share/applications/io.github.themetime.ThemeTime.desktop
install -Dm644 assets/icons/io.github.themetime.ThemeTime.svg \
  ~/.local/share/icons/hicolor/scalable/apps/io.github.themetime.ThemeTime.svg
```

Keep both user binaries together and ensure `~/.local/bin` is available to the
graphical session.

## 5. Install the background scheduler

After configuring and manually testing at least one phase:

```sh
themetime install-user-service
systemctl --user status themetime.service
```

This step needs no root access. Continue with
[Your first schedule](Your-First-Schedule).

## Optional components

Video wallpaper phases require Smart Video Wallpaper Reborn for Plasma 6 with
this exact plugin ID:

```text
luisbocanegra.smart.video.wallpaper.reborn
```

SDDM and Plymouth scheduling requires a separate administrator installation.
Only follow [Privileged themes](Privileged-Themes) if you need those actions.

For distribution-level prerequisite details and build behavior, use the full
[getting started guide](../docs/getting-started.md).
