# Architecture

ThemeTime separates presentation, schedule calculation, desktop integration,
and privileged operations. This keeps solar logic testable without KDE and keeps
the root surface much smaller than the desktop feature set.

## Runtime overview

```text
┌──────────────── Solar Observatory (Wails) ────────────────┐
│ JavaScript/CSS UI ⇄ bound Go App methods ⇄ config/inventory│
└────────────────────────────┬───────────────────────────────┘
                             │ config.json
                ┌────────────▼────────────┐
                │ scheduler + user daemon │
                └────────────┬────────────┘
                             │ ordered actions
                ┌────────────▼────────────┐
                │ KDE command/script layer│
                └─────────────────────────┘

config.json ── filter/export ── Polkit rootctl ── root schedule
                                                   │
                                           root daemon/applier
                                                   │
                                              SDDM/Plymouth
```

## Components

### Configuration and model

`internal/model` defines the versioned JSON data model, trigger/action enums,
defaults, and structural validation. `internal/config` owns XDG path resolution,
atomic saves, default creation, and file snapshots.

The model has no KDE or UI dependency, so the CLI, GUI, user daemon, and root
daemon all validate the same schema.

### Solar calculation and scheduling

`internal/solar` calculates daily solar events from coordinates and date.
`internal/scheduler` converts every enabled phase trigger into a timezone-aware
transition, applies event fallbacks and offsets, sorts transitions, and resolves
the active and next phases.

Current resolution examines yesterday, today, and tomorrow. This avoids a gap
after midnight and correctly identifies tomorrow's next phase. Phase JSON is
hashed to detect meaningful schedule changes without reapplying on every poll.

### KDE integration

`internal/kde` contains:

- an injectable command runner used by production code and tests;
- installed-theme and media discovery;
- ordered action application;
- Plasma shell script generation for per-screen wallpapers;
- snapshots before selected configuration-file writes.

Command arguments are passed directly except for `customCommand`, whose explicit
contract is `/bin/sh -c`. Wallpaper scripts JSON-encode interpolated values
before sending them to Plasma's `evaluateScript` method.

### User daemon

`internal/daemon` is a polling state machine. It reloads the config on every
tick, exits early when scheduling is disabled or effective state already
matches, and records state only after all attempted ordinary actions succeed.
The scheduler composes persistent actions by track and the daemon applies only
changed tracks. Privileged actions are skipped here by design.

The CLI hosts this daemon so the systemd service and foreground debugging use
the same implementation.

### Desktop interface and tray

`cmd/themetime-wails` embeds the production frontend at compile time. A bound Go
`App` exposes state loading, validated config saves, manual phase application,
location previews, diagnostics, and user-service installation. A mutex protects
shared config, path, inventory, and Wails context state.

The frontend in `cmd/themetime-wails/frontend/src` renders the Solar Observatory
and calls those bound methods. Preview is non-persistent; save is validated and
atomic. The window has a single-instance lock, hides on close, and uses a native
system tray loop. A second launch raises the existing window.

The tray queries the same scheduler for current/next labels and applies the
current phase through the same daemon helper as the UI.

### Privileged path

`internal/privileged`, `cmd/themetime-rootctl`, and `cmd/themetime-rootd` form a
separate data path. The user CLI exports a filtered schedule over standard input
to a Polkit-authorized fixed executable. The installer validates and atomically
writes that schedule. The root daemon revalidates, resolves it with the shared
scheduler, and dispatches only hard-coded SDDM/Plymouth operations.

Per-action root state is saved after each success. That detail is important when
Plymouth succeeds but a later action fails: the expensive image rebuild is not
repeated on the next poll.

## Data ownership

| Data | Writer | Readers |
| --- | --- | --- |
| User config | GUI, config initializer, user editor | GUI, CLI, user daemon, export command |
| User state | User daemon | User daemon |
| User snapshots | KDE applier | Administrator/user recovery |
| Privileged schedule | Root control helper | Root daemon |
| Root state | Root daemon | Root daemon |
| Root snapshots | Root applier | Administrator recovery |

The GUI process and user daemon can run concurrently. Atomic config replacement
prevents the daemon from observing a partially written GUI save.

## Failure behavior

- Invalid configuration stops the current operation before actions run.
- A failed user action leaves the previous state fingerprint in place, allowing
  retry on the next poll.
- Missing optional refresh utilities do not fail actions that have already
  written their primary configuration.
- Missing required commands fail their action with a direct diagnostic.
- Root schedules are validated both before installation and at load time.
- Neither daemon performs automatic rollback. Snapshots make manual recovery
  possible for the files ThemeTime directly changes.

## Trust boundaries

The user config is trusted as the current user. It may contain shell code via
`customCommand`, so importing it is equivalent to importing a script.

The root path treats the user config as untrusted. Filtering, whitelisting,
character restrictions, installed-theme checks, fixed helper paths, and no shell
execution prevent user-session capabilities from crossing into the privileged
daemon.
