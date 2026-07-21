# System tray and background service

The GUI tray and the scheduler daemon solve different problems:

| Component | Runs when | Purpose |
| --- | --- | --- |
| `themetime-wails` | You launch the app | Edit/preview schedules, tray controls |
| `themetime daemon` | User systemd starts it | Apply phases even when the GUI is closed |

Quitting one does not automatically stop the other.

## Tray behavior

Closing the main window hides it. Open the tray menu to see current and next
phase labels, show the window, apply the current phase, or quit the GUI.

The app uses a single-instance lock. Launching it again should bring the existing
window forward instead of starting another editor.

## Install automatic scheduling

```sh
themetime install-user-service
```

The installer writes an absolute binary path and starts the unit. Check it:

```sh
systemctl --user status themetime.service
```

Install without starting:

```sh
themetime install-user-service --now=false
```

Move the binary? Run the installer again so `ExecStart` is updated.

## Operate the service

```sh
systemctl --user start themetime.service
systemctl --user stop themetime.service
systemctl --user restart themetime.service
systemctl --user disable --now themetime.service
journalctl --user -u themetime.service -f
```

The daemon reloads config each poll; a normal GUI save does not need a restart.
Stopping it is useful while hand-editing JSON.

## Understand when it applies

The daemon stores the last transition and fingerprints for the composed action
tracks. It applies when:

- the active phase changes;
- an effective action track changes;
- the service starts and `reapplyOnStartup` is true;
- previous application failed and state was not finalized.

It does not apply merely because another 15-second poll occurred, and a new
video does not cause an unchanged theme track to be applied again.

## Run without systemd

One check and exit:

```sh
themetime daemon --once
```

Foreground loop with a different interval:

```sh
themetime daemon --poll 30s
```

This is useful for debugging journal/environment differences. See the full
[service reference](../docs/services-and-privileged-actions.md#user-daemon).
