# Developer guide

## Start with a clean validation run

```sh
make docs-check
make test
make build
```

`make docs-check` validates local links, heading anchors, code fences, and JSON
examples. `make test` includes that check, builds the Vite frontend, and runs
every Go package. `make build` produces all four executables.

## Follow a change through the layers

Schema or scheduling changes usually touch:

```text
internal/model → internal/scheduler → Wails view model → frontend → docs/tests
```

Desktop action changes usually touch:

```text
internal/model → internal/kde → inventory/UI options → docs/tests
```

Privileged changes additionally require root validation, state, Polkit/security,
and negative tests. Do not widen the root service to support arbitrary commands
or paths.

## Useful iteration commands

```sh
go test ./internal/scheduler
go test ./internal/kde
npm --prefix cmd/themetime-wails/frontend run build
go run ./cmd/themetime doctor
go run ./cmd/themetime daemon --once
```

The frontend production output is embedded in the Wails executable, so rebuild
it after JavaScript or CSS changes.

## Core invariants

- Config is validated before save or application.
- User config saves and root schedule installation are atomic.
- Phase resolution uses the configured timezone and spans three local dates.
- User failures do not advance the final phase fingerprint.
- The user applier skips privileged actions.
- The root loader accepts only SDDM/Plymouth theme identifiers.
- Preview operations do not mutate the persisted configuration.
- A global theme applies before more specific appearance overrides.

## Read next

- [Full development guide](../docs/development.md)
- [Architecture](../docs/architecture.md)
- [Configuration schema](../docs/configuration.md)
- [Action contracts](../docs/actions.md)

When behavior changes, update the canonical page in `docs/` and the relevant
task page in `wiki/` in the same change.
