# Packaging and releases

ThemeTime publishes a native Linux archive because the Wails interface depends
on GTK and WebKitGTK. The first supported release target is Linux x86-64 with
KDE Plasma 6.

## Install a release archive

Download the `.tar.gz` file and adjacent `.sha256` file from the GitHub release,
then verify and extract them:

```sh
sha256sum --check themetime-0.1.0-linux-x86_64.tar.gz.sha256
tar -xzf themetime-0.1.0-linux-x86_64.tar.gz
cd themetime-0.1.0-linux-x86_64
./install.sh --user --with-service
```

Without `--with-service`, the installer adds the binaries, desktop entry, and
icon but does not enable the background scheduler. Configuration remains in the
normal XDG directories.

The release binary still requires compatible GTK 3, WebKitGTK 4.1, Ayatana App
Indicator, and KDE Plasma libraries on the destination system. Run
`themetime doctor` after installation.

## System and privileged installation

A system installation places the normal binaries under `/usr/local`:

```sh
sudo ./install.sh --system
```

The restricted Polkit helper and root daemon remain opt-in:

```sh
sudo ./install.sh --system --with-privileged --enable-rootd
```

Do not enable privileged components if the schedule only changes user-session
themes and wallpapers.

## Uninstall

Run the uninstaller from an extracted copy of the same release:

```sh
./uninstall.sh --user
sudo ./uninstall.sh --system
```

Both modes preserve configuration and state. Add `--purge-config` only when
those files should also be permanently removed.

## Build a release locally

The project version is stored in `VERSION`. Build a deterministic archive and
checksum with:

```sh
make test
make package VERSION=0.1.0
(cd dist && sha256sum --check ./*.sha256)
```

For the complete local release gate, use the equivalent shortcut:

```sh
make release-check
```

The archive contains the two user binaries, two optional privileged binaries,
desktop integration, install/uninstall scripts, reference docs, wiki, license,
and changelog. `SOURCE_DATE_EPOCH` controls archive timestamps for reproducible
release builds.

## Publish from GitHub

1. Update `VERSION` and `CHANGELOG.md`.
2. Run `make test` and `make package` locally.
3. Commit the release changes.
4. Create and push a matching annotated tag:

   ```sh
   git tag -a v0.1.0 -m "ThemeTime 0.1.0"
   git push origin main v0.1.0
   ```

The release workflow tests the repository, builds on Ubuntu 24.04, creates the
x86-64 archive and SHA-256 checksum, and publishes both to a GitHub release.
GitHub also provides source archives for the tag.

## Version metadata

`make build` injects the value from `VERSION` and the current short commit into
the CLI:

```sh
themetime version
```

Release workflows override both values from the pushed tag and commit. Build
paths are removed from binaries with Go's `-trimpath` flag.
