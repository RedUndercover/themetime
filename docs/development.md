# Development guide

## Repository layout

```text
cmd/themetime/                 CLI and user daemon entry point
cmd/themetime-wails/           Wails API, view models, window, and tray
cmd/themetime-wails/frontend/  Modular Vite frontend and embedded build
cmd/themetime-rootctl/         Privileged schedule installer entry point
cmd/themetime-rootd/           Privileged scheduler entry point
internal/model/                Schema, enums, validation, defaults
internal/config/               Paths, loading, atomic saves, snapshots
internal/jsonfile/             Private atomic JSON persistence
internal/solar/                Solar event calculations
internal/scheduler/            Transition and current/next resolution
internal/kde/                  Discovery and desktop action application
internal/daemon/               User polling daemon and state
internal/privileged/           Filtered root schedule and actions
internal/doctor/               Environment diagnostics
internal/systemd/              User unit installation
assets/                        Desktop, systemd, and Polkit files
docs/                          Maintainer/reference documentation
docs/how-to.md                 Task-oriented workflows and recipes
.github/                       CI, releases, and contribution templates
```

## Build

The standard build is:

```sh
make build
```

It first runs:

```sh
npm --prefix cmd/themetime-wails/frontend ci --include=dev
npm --prefix cmd/themetime-wails/frontend run build
```

Then it builds the CLI, Wails app, root helper, and root daemon. The Wails binary
uses `production,desktop` build tags and adds `webkit2_41` when
`pkg-config --exists webkit2gtk-4.1` succeeds.

Because `frontend/dist` is embedded by Go, rebuild the frontend before compiling
the Wails binary after JavaScript or CSS changes. `make build` always does this.

Create a local snapshot of every distributable artifact with `make package`.
This runs the pinned GoReleaser v2 release pipeline and writes an Arch package,
portable archive, and checksum manifest under `dist/`. See
[Packaging and releases](packaging-and-releases.md) for prerequisites,
artifact names, versioning, and tag instructions.

For CLI-only iteration:

```sh
go run ./cmd/themetime doctor
go run ./cmd/themetime daemon --once
```

For frontend-only layout iteration:

```sh
npm --prefix cmd/themetime-wails/frontend test
npm --prefix cmd/themetime-wails/frontend run dev
```

The browser dev server does not provide Wails Go bindings by itself, so backend
calls need a Wails runtime or temporary development fixtures. The supported
end-to-end check is the production frontend build plus the Go application.

## Tests and formatting

Run the full project test target:

```sh
make test
```

This validates the documentation, runs the frontend unit tests and production
build, and tests all Go packages. To check documentation only or iterate on Go
tests:

```sh
make docs-check
go test ./...
go test ./internal/scheduler
go test ./internal/kde -run TestName
```

Format Go sources with:

```sh
make fmt
```

Frontend tests use Node's built-in test runner. Keep schedule/config mutations
and formatting in `src/domain.js` so they remain testable without a browser;
`src/views.js` owns markup and `src/main.js` owns Wails orchestration and the
delegated event handlers. Continue to manually exercise save, location preview,
phase editing, sheet focus behavior, and tray lifecycle for interface changes.

## Testing design

Scheduler tests should use explicit timestamps and IANA zones so daylight-saving
behavior is reproducible. Include previous-day and next-day boundaries when
changing active/next resolution.

KDE and privileged tests use the command runner abstraction; do not invoke the
real desktop or modify `/etc` in unit tests. Test command names and arguments,
missing-command behavior, generated Plasma scripts, action ordering, validation,
and state deduplication.

Configuration tests should cover invalid enums and boundaries as well as a
round-trip through JSON. Privileged validation needs negative tests for every
new field that could widen the root surface.

## Add a trigger

1. Add the identifier and labels to `TriggerKind` and `TriggerDefinitions` in
   `internal/model`; the frontend picker reads this metadata from `UIState`.
2. Extend trigger validation.
3. Resolve the event or clock in `internal/scheduler` and define unavailable
   behavior.
4. Add any trigger-specific frontend icon or marker presentation.
5. Add model, scheduler, GUI, and timezone-boundary tests.
6. Update [Configuration](configuration.md), the
   [how-to guide](how-to.md), and example schedules.

Adding a trigger changes both persisted configuration and presentation, even
when the schema version stays the same.

## Add a user action

1. Add the `ActionType` constant and label in `internal/model`.
2. Define its validation contract.
3. Implement it in `internal/kde.Applier.ApplyAction` using argument-based
   command execution where possible.
4. Assign its ordering group.
5. Extend inventory discovery and GUI action options if it is selectable.
6. Add success, missing dependency, escaping, and failure tests.
7. Document its `value`, `screen`, and `values` contract in
   [Actions](actions.md).

Snapshot any user configuration file before ThemeTime writes it. Avoid treating
display names as stable package IDs.

## Change a privileged action

Privileged changes require a higher review bar. Update filtering, schedule
validation, root application, installed-resource checks, Polkit assumptions,
state behavior, tests, and the security documentation together.

Never route `customCommand`, wallpaper scripts, arbitrary paths, or arbitrary
program execution through the root daemon. Keep the accepted input representable
as a small whitelist of identifiers.

## GUI/backend contract

The Wails backend methods exposed to JavaScript are:

```text
GetState
SaveConfig
ApplyPhase
PreviewLocation
RunDoctor
InstallUserService
```

`UIState` is the view model shared across the initial render and refreshes. Its
`triggerOptions` field is the frontend source of truth for trigger ordering and
labels. When changing its JSON shape, update frontend reads in the same commit.
Keep previews side-effect free and validate again in the Go save method rather
than relying on browser controls.

## Documentation maintenance

Reference behavior belongs in its dedicated page under `docs/`; common
workflows and recipes belong in `docs/how-to.md`. When commands, schema fields,
paths, action values, fallback behavior, or service security change, update
both in the same change. Run `make docs-check` before merging; it verifies local
links and anchors, balanced code fences, and every fenced JSON example.
