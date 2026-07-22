# Contributing to ThemeTime

Thank you for helping improve ThemeTime. Start with the
[development guide](docs/development.md) and the
[architecture reference](docs/architecture.md).

## Development workflow

1. Create a focused branch from `main`.
2. Keep scheduler calculations deterministic and preserve the privileged
   action whitelist.
3. Add or update tests for behavioral changes.
4. Update the canonical reference and any affected workflow in
   `docs/how-to.md`.
5. Run:

   ```sh
   make test
   make build
   ```

6. Open a pull request using the repository template.

Avoid committing binaries, frontend dependencies, generated `dist/` output,
local configuration, logs, or wallpaper media. Never include secrets or private
commands from a real ThemeTime configuration in an issue or test fixture.

## Commit and pull request scope

Prefer small commits that explain why behavior changed. A pull request should
cover one coherent outcome and call out changes to scheduling boundaries,
desktop commands, persisted JSON, or privileged behavior.

By contributing, you agree that your contribution is licensed under the MIT
License included with this repository.
