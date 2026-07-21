# Troubleshooting

Start with:

```sh
themetime doctor
themetime show-config
systemctl --user status themetime.service
```

Then use the symptom below. Commands assume an installed binary; substitute
`./bin/themetime` in the source tree.

## The GUI does not start

Run the Wails binary directly to expose its error:

```sh
themetime-wails
```

If `themetime gui` says the GUI was not found, keep `themetime` and
`themetime-wails` in the same directory, put `themetime-wails` on `PATH`, or set:

```sh
THEMETIME_GUI=/absolute/path/to/themetime-wails themetime gui
```

If diagnostics report WebKitGTK missing, install the development/runtime package
for WebKitGTK 4.1 or 4.0 supplied by the distribution and rebuild. A blank web
view is usually a WebKitGTK or embedded frontend build issue; run:

```sh
make build-wails-frontend
make build
```

## Closing the window does not exit

This is expected. Closing hides ThemeTime to the system tray. Use **Quit** from
the tray menu. If the tray is not visible, start `themetime-wails` in a terminal
to inspect tray errors, or terminate that GUI process normally. The systemd user
daemon is separate and may remain active.

## Solar times are wrong

Check all location fields, especially signs and timezone:

- latitude: north positive, south negative;
- longitude: east positive, west negative;
- timezone: an IANA identifier, not a numeric offset.

Preview the location in the GUI before saving. Compare `themetime show-config`
with the expected coordinates. If markers show fallback-like round times at a
high latitude, the event may not occur on that date; review
[polar fallbacks](configuration.md#polar-and-unavailable-event-fallbacks).

## The wrong phase is active after midnight

ThemeTime intentionally carries the final phase from yesterday until today's
first transition. Inspect every enabled phase and timezone. Duplicate or
near-identical transition times make a schedule harder to reason about; move one
trigger or disable the unintended phase.

## The service is installed but nothing changes

Inspect the journal:

```sh
journalctl --user -u themetime.service -n 100 --no-pager
```

Then check:

1. `runtime.enabled` is `true`.
2. At least one enabled phase resolves.
3. The service's `ExecStart` path still exists.
4. Required KDE commands are installed.
5. The service runs inside the expected graphical user session.
6. Action values are package IDs or absolute media paths, not display labels.

Test the same code path interactively:

```sh
themetime daemon --once
themetime apply --phase <phase-id>
```

Reinstall the unit after moving the binary:

```sh
themetime install-user-service
```

## The daemon repeatedly applies a phase

Repeated application normally means at least one action is failing, so state is
not finalized. Search the journal for `apply ... failed`. Fix the missing
command, invalid theme ID, inaccessible file, or failing custom command.

With `reapplyOnStartup: true`, one application after every service restart is
expected. Config edits also change the phase fingerprint and trigger one new
application.

## A manual JSON edit is rejected

Use:

```sh
themetime show-config
```

Common validation failures include:

- invalid JSON syntax or trailing commas;
- a clock not exactly in 24-hour `HH:MM` form;
- a solar trigger that still contains a `clock` field;
- duplicate or empty phase IDs;
- missing action values;
- latitude/longitude outside their ranges;
- an empty timezone.

An unknown non-empty timezone passes structural config validation but fails when
the GUI preview or scheduler tries to load it. Use an IANA name available in the
host timezone database and check the daemon journal for the load error.

Restore the manual backup or correct the reported field before restarting the
service. The GUI will not overwrite an invalid config it cannot load.

## A color, icon, font, cursor, or decoration does not change

Run `themetime doctor` and verify the action's command. KDE package display names
may differ from their IDs, so inspect installed choices in the GUI.

Some applications cache fonts and icons. Restart the affected app, then Plasma
or the session if needed. Window-decoration refresh uses `qdbus6` when available;
without it, the config can be correct but visible only after KWin reloads.

Snapshots of `kdeglobals` or `kwinrc` may be available under
`~/.local/state/themetime/snapshots/` if manual recovery is required.

## A wallpaper does not change

Use an absolute readable path or a leading `~/` path. ThemeTime does not expand
`$HOME` or other shell expressions. For static wallpapers, check
`plasma-apply-wallpaperimage`; per-screen and fallback behavior requires `qdbus6`
and a running Plasma shell.

If `screen` is set, confirm the current Plasma screen identifier. Remove the
field to target all desktops. Verify the extension is supported and that the
service user can read the file.

## A video wallpaper does not play

Confirm all three prerequisites:

```sh
themetime doctor
command -v qdbus6
kpackagetool6 --type Plasma/Wallpaper --list | \
  grep luisbocanegra.smart.video.wallpaper.reborn
```

Install the Plasma 6 version of Smart Video Wallpaper Reborn if its exact plugin
ID is absent. Use an absolute AVI, M4V, MKV, MOV, MP4, or WebM path. Plugin
configuration option codes may change across plugin releases; start with
ThemeTime's defaults before adding advanced `values`.

## A custom command works in a terminal but not in the daemon

Systemd user services often have a smaller `PATH` and different environment.
Use absolute executable paths, redirect output explicitly, and avoid relying on
interactive shell startup files. The action runs through `/bin/sh`, not
necessarily Bash.

Test with the phase command and inspect both JSON output and the journal:

```sh
themetime apply --phase <phase-id>
journalctl --user -u themetime.service -n 50 --no-pager
```

## Privileged actions never run

Saving the user config does not update the root schedule. After every privileged
change:

```sh
themetime install-privileged-schedule
sudo /usr/local/libexec/themetime-rootd --once
```

Also check:

```sh
sudo systemctl status themetime-rootd.service
sudo journalctl -u themetime-rootd.service -n 100 --no-pager
sudo test -r /etc/themetime/privileged-schedule.json
```

The requested SDDM or Plymouth ID must already be installed and contain only
safe identifier characters. The helper must remain at the exact path named by
the Polkit policy.

## Plymouth is rebuilt repeatedly

Check whether `/var/lib/themetime/state.json` is writable and valid, and inspect
the root journal for another action failing after Plymouth. Current root state
records each successful action before completing the phase specifically to avoid
repeat rebuilds. Do not delete the state file as a routine fix.

## Collect diagnostics for a bug report

Capture:

```sh
themetime doctor
themetime show-config
systemctl --user status themetime.service --no-pager
journalctl --user -u themetime.service -n 100 --no-pager
```

For privileged problems also capture root service status and journal entries.
Review config and logs before sharing them: custom command text, usernames,
paths, coordinates, and environment details may be sensitive.
