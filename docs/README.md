# ThemeTime documentation

This directory contains the authoritative reference documentation for the
current ThemeTime source tree. The [wiki](../wiki/Home.md) reorganizes the same
material around common tasks and is suitable for publishing as a GitHub wiki.

## Start here

- [Getting started](getting-started.md) covers prerequisites, building,
  installation, and the first successful run.
- [User guide](user-guide.md) explains the Solar Observatory interface and the
  normal schedule workflow.
- [Troubleshooting](troubleshooting.md) is the fastest route when an action or
  service does not work.

## Reference

- [Configuration](configuration.md) documents every JSON field, trigger, file,
  fallback, and validation rule.
- [Actions](actions.md) documents every action type and its optional values.
- [CLI](cli.md) lists every command and flag.
- [Services and privileged actions](services-and-privileged-actions.md) covers
  the user daemon, system daemon, Polkit boundary, state, and recovery.
- [Packaging and releases](packaging-and-releases.md) covers release archives,
  installers, checksums, versioning, and GitHub automation.

## Contributors

- [Architecture](architecture.md) follows data from the GUI through scheduling
  and application.
- [Development](development.md) covers the source layout, frontend workflow,
  builds, tests, and extension points.

## Documentation conventions

- Commands beginning with `./bin/` assume `make build` was run in the project
  root. Installed users can substitute `themetime`.
- Paths beginning with `~/.config` and `~/.local/state` are defaults. XDG
  environment variables take precedence.
- Action and trigger identifiers are written exactly as they appear in JSON.
- Examples use JSON because it is ThemeTime's persisted configuration format.

If behavior and prose ever disagree, treat the validation and application code
as authoritative and update these documents in the same change.

Validate local links, heading anchors, fenced code blocks, and JSON examples:

```sh
make docs-check
```
