# Getting started

## Platform and prerequisites

ThemeTime is intended for a KDE Plasma 6 desktop on Linux. The scheduler itself
is written in Go; the Solar Observatory desktop interface is a Wails application
whose frontend is built with Vite.

To build the full application, install:

- Go 1.25 or newer;
- Node.js and npm;
- `make`, a C compiler, and `pkg-config`;
- WebKitGTK development files, preferably WebKitGTK 4.1;
- Plasma 6 command-line tools.

At runtime, the core user-session integration expects these commands:

```text
plasma-apply-colorscheme
kwriteconfig6
kreadconfig6
qdbus6
```

Depending on the actions you use, ThemeTime can also use:

```text
plasma-apply-lookandfeel
plasma-apply-desktoptheme
plasma-apply-cursortheme
plasma-apply-wallpaperimage
kpackagetool6
pkexec
```

Package names vary between distributions. `themetime doctor` checks commands,
the current KDE session, WebKitGTK, the video wallpaper plugin, service files,
and the root helper, then prints targeted hints.

## Build from source

Clone or enter the source tree, then run:

```sh
make build
```

This performs `npm ci` and a production frontend build before compiling all
four Go executables into `bin/`. WebKitGTK 4.1 support is selected automatically
when `pkg-config` can find `webkit2gtk-4.1`; otherwise the build uses the Wails
default WebKitGTK 4.0 integration.

Verify the environment:

```sh
./bin/themetime doctor
```

## Run the interface

```sh
./bin/themetime gui
```

The CLI looks for a sibling `themetime-wails` executable, so keep those two
binaries together. You may also run the desktop executable directly:

```sh
./bin/themetime-wails
```

On the first start, ThemeTime creates a validated default configuration at
`~/.config/themetime/config.json` unless `XDG_CONFIG_HOME` is set. The default
location is New York and the schedule has morning and evening phases.

Closing the window hides it to the system tray. Use the tray menu to show the
window again, apply the current phase, or quit.

## Install for one user

Install both binaries, the desktop launcher, and its matching solar icon:

```sh
make install-user-assets
```

The equivalent manual installation is:

```sh
install -Dm755 bin/themetime ~/.local/bin/themetime
install -Dm755 bin/themetime-wails ~/.local/bin/themetime-wails
install -Dm644 assets/desktop/io.github.themetime.ThemeTime.desktop \
  ~/.local/share/applications/io.github.themetime.ThemeTime.desktop
install -Dm644 assets/icons/io.github.themetime.ThemeTime.svg \
  ~/.local/share/icons/hicolor/scalable/apps/io.github.themetime.ThemeTime.svg
```

Ensure `~/.local/bin` is in the graphical session's `PATH`, then refresh the
application database if your desktop does not notice the entry automatically.
The icon and desktop-file IDs intentionally match the Wails window ID so Plasma
groups the running window with its launcher and uses the same solar icon.

Install the user scheduler using the installed CLI:

```sh
themetime install-user-service
```

The generated unit contains the executable's absolute path and is enabled and
started by default. Re-run this command after moving or replacing the CLI at a
different path.

## Create the first custom schedule

1. Open **Location**, enter coordinates and an IANA timezone, and use preview to
   check the calculated solar events.
2. Select the morning phase, click the sunrise marker, and set an offset.
3. Add a color scheme or wallpaper action.
4. Select the evening phase and repeat with sunset.
5. Save the schedule.
6. Use **Apply now** to test a phase without waiting for its trigger.

The background daemon reloads the config on each polling cycle, so saved changes
take effect without restarting the service.

## Optional integrations

For video wallpapers, install Smart Video Wallpaper Reborn for Plasma 6. Its
plugin ID must be `luisbocanegra.smart.video.wallpaper.reborn`.

For scheduled SDDM or Plymouth themes, finish the separate privileged setup in
[Services and privileged actions](services-and-privileged-actions.md). Do not
install the root components if you only need desktop-session actions.

## Next steps

- Learn the interface in the [user guide](user-guide.md).
- Review all action capabilities in the [action reference](actions.md).
- Use the [first-schedule wiki walkthrough](../wiki/Your-First-Schedule.md) for a
  shorter task checklist.
