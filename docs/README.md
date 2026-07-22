# ThemeTime documentation

These pages are the canonical documentation for the current ThemeTime source
tree.

## Install and use

- [Getting started](getting-started.md) covers release installation, source
  builds, the first run, the user service, upgrades, and removal.
- [User guide](user-guide.md) explains the Solar Observatory interface and the
  normal schedule workflow.
- [How-to guide](how-to.md) provides a first-schedule walkthrough, solar-event
  guidance, action examples, and advanced configuration recipes.
- [Troubleshooting](troubleshooting.md) starts with a short diagnostic sequence
  and continues with symptom-specific guidance.

## Reference

- [Configuration](configuration.md) documents every persisted field, trigger,
  fallback, and validation rule.
- [Actions](actions.md) documents every action type and optional value.
- [CLI](cli.md) lists every command and flag.
- [Services and privileged actions](services-and-privileged-actions.md) covers
  the user daemon, system daemon, Polkit boundary, state, and recovery.

## Maintainers

- [Architecture](architecture.md) follows data from the GUI through scheduling
  and application.
- [Development](development.md) covers the source layout, builds, tests, and
  extension points.
- [Packaging and releases](packaging-and-releases.md) covers local release
  validation and publishing tagged GitHub releases.

## Conventions

- Commands beginning with `./bin/` assume `make build` was run in the project
  root. Installed users can substitute `themetime`.
- XDG environment variables override paths shown under `~/.config` and
  `~/.local/state`.
- Action and trigger identifiers match their JSON representation exactly.
- Behavior in validation and application code takes precedence over prose.

Validate local links, heading anchors, code fences, and JSON examples with:

```sh
make docs-check
```
