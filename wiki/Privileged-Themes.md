# Privileged themes

Only SDDM login themes and Plymouth boot themes use root privileges. Ordinary
desktop actions and custom commands never cross this boundary.

## Before you begin

Confirm the desired IDs are installed:

```sh
find /usr/share/sddm/themes /usr/local/share/sddm/themes \
  -mindepth 1 -maxdepth 1 -type d 2>/dev/null
plymouth-set-default-theme --list
```

Allowed IDs contain only letters, digits, dot, underscore, and hyphen, with no
`..` sequence.

## Install the restricted root service

From the source tree:

```sh
make build
sudo make install-root-assets
sudo systemctl enable --now themetime-rootd.service
```

This installs fixed helpers, a systemd service, and a Polkit policy. The policy
authorizes only `/usr/local/libexec/themetime-rootctl`.

## Add actions

Add these to the desired phases:

```json
{ "type": "sddmTheme", "value": "breeze" }
```

```json
{ "type": "plymouthTheme", "value": "bgrt" }
```

Use installed IDs, not labels or paths.

## Export after every relevant edit

```sh
themetime install-privileged-schedule
```

Authenticate through Polkit. This copies only privileged actions and their
schedule context into `/etc/themetime/privileged-schedule.json`. It is a snapshot,
not a live reference to the user config.

Test immediately:

```sh
sudo /usr/local/libexec/themetime-rootd --once
sudo journalctl -u themetime-rootd.service -n 50 --no-pager
```

## What application does

For SDDM, ThemeTime verifies the installed directory and writes
`/etc/sddm.conf.d/90-themetime.conf`. It does not restart SDDM, because that can
end the current desktop session.

For Plymouth, it verifies the theme list and runs
`plymouth-set-default-theme -R`, which rebuilds the boot image. Root-side state
prevents repeating a successful rebuild every minute.

## Security summary

The export path filters out every other action, validates the schema and user
ID, restricts theme identifiers, uses a fixed authenticated helper, and validates
again when root loads the schedule. There is no privileged shell or arbitrary
file action.

For installation paths, recovery, and root state details, read the full
[services and privileged actions reference](../docs/services-and-privileged-actions.md).
