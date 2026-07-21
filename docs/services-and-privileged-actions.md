# Services and privileged actions

ThemeTime has two independent schedulers:

```text
user config ──> user daemon ──> KDE session actions
      │
      └── explicit export via Polkit ──> root schedule ──> SDDM/Plymouth
```

The normal installation only needs the user daemon. The root path is optional,
narrowly validated, and cannot run custom commands.

## User daemon

Install the user unit:

```sh
themetime install-user-service
```

This creates `~/.config/systemd/user/themetime.service` with:

- the absolute path to the selected `themetime` binary;
- `themetime daemon` as `ExecStart`;
- ordering after the graphical session and Plasma shell;
- restart on failure after five seconds;
- enablement under `default.target`.

The installer runs `systemctl --user daemon-reload` and, unless
`--now=false` is supplied, `systemctl --user enable --now
themetime.service`.

Common operations:

```sh
systemctl --user status themetime.service
journalctl --user -u themetime.service -f
systemctl --user restart themetime.service
systemctl --user disable --now themetime.service
```

The daemon polls every 15 seconds by default. Each tick reloads and validates
the config, resolves the active transition, composes the latest persistent value
for every action track, compares those tracks with saved state, and applies only
what changed. It logs failed and skipped actions to the systemd journal.

### User state

State is written with mode `0600` to:

```text
$XDG_STATE_HOME/themetime/state.json
```

or `~/.local/state/themetime/state.json` by default. It stores the last phase
ID, transition time, effective-state fingerprint, per-action fingerprints, and
application time. A video transition therefore changes the wallpaper fingerprint
without making ThemeTime reapply an unchanged color scheme.

With `runtime.reapplyOnStartup: true`, the first daemon tick reapplies every
persistent effective action. Historical custom commands are not replayed. Failed
actions do not update state, so the changed tracks are retried.

## Why the root scheduler is separate

SDDM controls the login screen and Plymouth controls the boot splash, so both
need system-level changes. Giving the normal daemon broad root access would also
give its custom command action root access. ThemeTime instead uses:

- a small installer invoked by `pkexec`;
- strict JSON validation and action filtering;
- a root daemon that implements only two hard-coded action types;
- fixed paths and installed-theme checks;
- root-side state to avoid repeated expensive changes.

The boundary accepts no arbitrary executable name, file path, shell expression,
or non-privileged action.

## Install root components

From the source tree:

```sh
make build
sudo make install-root-assets
sudo systemctl enable --now themetime-rootd.service
```

Build as your normal user first. The privileged Make target then installs:

| File | Installed path |
| --- | --- |
| Root control helper | `/usr/local/libexec/themetime-rootctl` |
| Root daemon | `/usr/local/libexec/themetime-rootd` |
| Polkit policy | `/usr/share/polkit-1/actions/io.github.themetime.rootctl.policy` |
| systemd system unit | `/etc/systemd/system/themetime-rootd.service` |

It then reloads the system systemd manager. The unit polls once a minute and
restarts after ten seconds if its process fails.

The Polkit policy requires administrator authentication for active, inactive,
and remote contexts and fixes the authorized executable path to
`/usr/local/libexec/themetime-rootctl`. It also marks the helper as non-GUI.

## Export the schedule

Add `sddmTheme` and/or `plymouthTheme` actions to user phases, then run:

```sh
themetime install-privileged-schedule
```

ThemeTime:

1. loads and validates the user config;
2. removes phases without privileged actions and removes all user actions from
   the remaining phases;
3. records the current numeric user ID and export time;
4. invokes the fixed root helper with `pkexec`;
5. validates again and atomically installs the result with mode `0600` at
   `/etc/themetime/privileged-schedule.json`.

Exporting is a snapshot operation. Saving the GUI or user JSON later does not
change the root copy. Re-export after changing location, runtime settings,
phase triggers, enabled state, IDs, or privileged action values.

### Validation boundary

An exported schedule is rejected when:

- it contains any action other than `sddmTheme` or `plymouthTheme`;
- a theme ID contains characters outside `A-Z`, `a-z`, `0-9`, `.`, `_`, and
  `-`;
- a theme ID contains `..`;
- its embedded ThemeTime config is invalid;
- the claimed user ID does not match the requested user ID.

The root daemon independently reloads and validates the installed file.

## SDDM behavior

Before applying an SDDM theme, ThemeTime checks for its directory beneath:

```text
/usr/share/sddm/themes
/usr/local/share/sddm/themes
```

It snapshots an existing `/etc/sddm.conf.d/90-themetime.conf` under
`/var/lib/themetime/snapshots/`, then writes:

```ini
[Theme]
Current=<theme-id>
```

The change normally appears at the next login screen start or refresh. ThemeTime
does not restart SDDM because doing so can terminate the active graphical
session.

## Plymouth behavior

ThemeTime requires `plymouth-set-default-theme`, verifies the requested ID in
its `--list` output, and runs:

```text
plymouth-set-default-theme -R <theme-id>
```

The rebuild can be slow and is distribution-specific. The root state file at
`/var/lib/themetime/state.json` stores per-action fingerprints while a phase is
being applied and a final phase fingerprint after success. This prevents a
successful Plymouth rebuild from being repeated merely because another action
in the same phase failed.

## Inspect and test the root daemon

```sh
sudo systemctl status themetime-rootd.service
sudo journalctl -u themetime-rootd.service -f
sudo /usr/local/libexec/themetime-rootd --once
```

The one-shot form is useful after exporting. It prints JSON when it applies
actions and returns an error when the schedule cannot be loaded or an action
fails.

Never hand-edit the installed privileged schedule as a normal workflow. Edit the
user config and export again so both validation layers run.

## Removal

Disable the services first:

```sh
systemctl --user disable --now themetime.service
sudo systemctl disable --now themetime-rootd.service
```

Removing installed files is deliberately left to the packaging or system
administrator. Removing ThemeTime's SDDM fragment can reveal another SDDM theme
configured elsewhere; removing a Plymouth package or setting requires the
distribution's Plymouth tooling.
