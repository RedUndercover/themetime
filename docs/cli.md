# Command-line reference

The `themetime` executable controls the GUI, user daemon, diagnostics, direct
phase application, and service installation.

```text
themetime <command> [options]
```

Running `themetime` with no command is equivalent to `themetime gui`.

## `gui`

```sh
themetime gui
```

Starts the Solar Observatory application. Resolution order is:

1. the path in `THEMETIME_GUI`, if set;
2. a `themetime-wails` executable beside the CLI;
3. `go run` from a ThemeTime source-tree working directory;
4. `themetime-wails` on `PATH`.

The source-tree fallback uses Wails build tags `production,desktop` and adds
`webkit2_41` when `pkg-config` finds WebKitGTK 4.1.

Example override:

```sh
THEMETIME_GUI=/opt/themetime/themetime-wails themetime gui
```

## `daemon`

```sh
themetime daemon [--once] [--poll 15s] [--config PATH]
```

Runs the user-session scheduler.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--once` | false | Resolve and apply once, then exit. |
| `--poll` | `15s` | Poll interval as a Go duration, such as `30s` or `2m`. |
| `--config` | XDG user config | Read an alternate configuration file. |

The installed user service runs this command continuously. SIGINT and SIGTERM
shut it down cleanly.

Useful checks:

```sh
themetime daemon --once
themetime daemon --once --config /tmp/test-schedule.json
themetime daemon --poll 30s
```

The alternate config affects configuration loading; scheduler state and
snapshots still use the normal user state paths.

## `apply`

```sh
themetime apply --phase <id>
```

Immediately applies the named phase from the normal user config, regardless of
its trigger or enabled state. `--phase` is required and matches the phase `id`,
not its display name.

```sh
themetime apply --phase evening
```

The command prints JSON results for each action. A result includes the action,
whether it was skipped, a message, and an optional error. Privileged actions are
skipped here because only the root daemon may apply them.

## `doctor`

```sh
themetime doctor
```

Reports:

- required and optional KDE commands;
- whether the process appears to be in a KDE session;
- WebKitGTK availability for the GUI;
- Smart Video Wallpaper Reborn installation;
- config location;
- user service installation;
- privileged helper installation.

Run this before opening a bug report and after changing desktop packages.

## `install-user-service`

```sh
themetime install-user-service [--now=true] [--binary PATH]
```

Writes `~/.config/systemd/user/themetime.service`, reloads the user systemd
manager, and normally enables and starts the service.

| Flag | Default | Meaning |
| --- | --- | --- |
| `--now` | true | Enable and start immediately. Use `--now=false` to install only. |
| `--binary` | current executable | Absolute or resolvable path for the service's CLI. |

Examples:

```sh
themetime install-user-service
themetime install-user-service --now=false
themetime install-user-service --binary /opt/themetime/themetime
```

The installer refuses to use the Wails executable as the daemon. If invoked
from `themetime-wails`, it looks for the real `themetime` CLI beside it or on
`PATH`.

## `install-privileged-schedule`

```sh
themetime install-privileged-schedule
```

Filters the user config down to `sddmTheme` and `plymouthTheme` actions, wraps it
with the current numeric user ID and export time, and sends it over standard
input to the fixed root helper through `pkexec`.

The helper must be installed and executable at:

```text
/usr/local/libexec/themetime-rootctl
```

This command replaces the installed privileged schedule. Re-run it whenever
privileged phases, triggers, actions, location, or runtime settings change.

## `show-config`

```sh
themetime show-config
```

Loads, validates, and prints the normal config path followed by pretty-printed
JSON. If the file does not exist, this command creates the default config first.

## `version`

```sh
themetime version
themetime --version
```

Prints the application version and, for repository or release builds, the short
source commit embedded by the Makefile.

This is useful as a read-only validation check after manual editing:

```sh
themetime show-config >/dev/null
```

## Help and errors

```sh
themetime help
themetime --help
themetime -h
```

Unknown commands and failed operations print an error to standard error and
exit nonzero. Individual Go flag sets also support their standard `-h` output.
