# Security policy

## Supported versions

Until ThemeTime reaches 1.0, security fixes are applied to the latest tagged
release and the `main` branch.

## Reporting a vulnerability

Do not open a public issue for a vulnerability involving the Polkit helper,
root daemon, command execution, path validation, or privilege boundaries. Use
[GitHub private vulnerability reporting](https://github.com/themetime/themetime/security/advisories/new).

Include the affected version, environment, reproduction steps, impact, and any
suggested mitigation. Remove credentials, private commands, and personal paths.

The root surface intentionally accepts only SDDM and Plymouth theme IDs.
Reports that demonstrate a way to route arbitrary commands or paths across that
boundary are especially important.
