# Getting started

ThemeTime targets Linux x86-64 with KDE Plasma 6. Choose a packaged release for
normal use or build from source for development.

## Install on Arch Linux

Download the `.pkg.tar.zst` file and `checksums.txt` from the same GitHub
release. Replace `VERSION` in these examples with the release number.

```sh
sha256sum --ignore-missing --check checksums.txt
sudo pacman -U themetime_VERSION_linux_amd64.pkg.tar.zst
systemctl --user enable --now themetime.service
themetime doctor
```

The package installs all ThemeTime executables, the desktop launcher and icon,
systemd units, the Polkit policy, and documentation. It does not enable the
optional privileged service.

## Run the portable archive

The portable archive is useful on other Linux distributions with compatible
runtime libraries:

```sh
sha256sum --ignore-missing --check checksums.txt
mkdir themetime-VERSION
tar -xzf themetime-VERSION-linux-x86_64.tar.gz -C themetime-VERSION
cd themetime-VERSION
./themetime doctor
./themetime gui
```

Keep `themetime` and `themetime-wails` together. The archive runs in place; it
does not add a desktop launcher or service. It requires GTK 3, WebKitGTK 4.1,
Ayatana App Indicator, and KDE Plasma libraries on the destination system.

## Build from source

Install:

- Go 1.25 or newer;
- Node.js and npm;
- `make`, a C compiler, and `pkg-config`;
- GTK 3 and WebKitGTK development files;
- KDE Plasma 6 command-line tools.

Then build and inspect the environment:

```sh
make build
./bin/themetime doctor
```

`make build` runs `npm ci`, builds the frontend, and compiles four executables
under `bin/`: the CLI/daemon, desktop interface, restricted Polkit helper, and
privileged daemon. WebKitGTK 4.1 is selected when `pkg-config` finds it;
otherwise the Wails default WebKitGTK 4.0 integration is used.

Start the interface:

```sh
./bin/themetime gui
```

The first run creates a validated configuration at
`~/.config/themetime/config.json` unless `XDG_CONFIG_HOME` is set. Closing the
window hides it to the system tray; use **Quit** in the tray menu to stop the
interface completely.

## Install from the source tree

Install the user binaries, launcher, and icon:

```sh
make install-user-assets
```

Ensure `~/.local/bin` is available to the graphical session. Install and start
the scheduler with the installed CLI:

```sh
themetime install-user-service
systemctl --user status themetime.service
```

The generated unit records the CLI's absolute path. Run the installer again if
the binary moves. Use `--now=false` to install the unit without starting it.

## Create and test a schedule

1. Open **Location**, enter decimal coordinates and an IANA timezone, then
   preview the calculated solar events.
2. Select a phase and choose a solar event or fixed clock time.
3. Add a color scheme, wallpaper, or another action.
4. Save the schedule and use **Apply now**.
5. Repeat for other parts of the day.

The daemon reloads the configuration on each polling cycle. Saved changes do
not require a service restart. See the [how-to guide](how-to.md) for a complete
day/night example and reusable recipes.

## Optional integrations

Video wallpapers require Smart Video Wallpaper Reborn for Plasma 6 with plugin
ID `luisbocanegra.smart.video.wallpaper.reborn`.

Scheduled SDDM and Plymouth themes require the separate privileged setup in
[Services and privileged actions](services-and-privileged-actions.md). Do not
install root components for ordinary desktop-session actions.

## Upgrade or remove

Upgrade the Arch package by installing the newer `.pkg.tar.zst` with
`pacman -U`. To remove it:

```sh
systemctl --user disable --now themetime.service
sudo systemctl disable --now themetime-rootd.service
sudo pacman -Rns themetime
```

The root-service command is necessary only if that optional service was
enabled. A portable installation is removed by deleting its extraction
directory. User configuration and state are preserved in both cases; remove
their XDG directories manually only when that data is no longer needed.
