# Configuration recipes

The GUI covers the common schedule workflow. Edit JSON for advanced values,
font profiles, shell commands, or per-screen targets.

## Safely edit the config

```sh
systemctl --user stop themetime.service
cp ~/.config/themetime/config.json ~/.config/themetime/config.json.manual-backup
$EDITOR ~/.config/themetime/config.json
themetime show-config
systemctl --user start themetime.service
```

If `show-config` reports an error, fix or restore the file before starting the
daemon. The default path changes when `XDG_CONFIG_HOME` is set.

## Create a fixed bedtime phase

Add a phase like:

```json
{
  "id": "bedtime",
  "name": "Bedtime",
  "color": "#202744",
  "enabled": true,
  "start": {
    "kind": "clock",
    "clock": "22:30"
  },
  "actions": [
    { "type": "colorScheme", "value": "BreezeDark" }
  ]
}
```

The bedtime phase remains active past midnight until another phase begins.

## Disable scheduling temporarily

Set:

```json
{
  "runtime": {
    "enabled": false,
    "reapplyOnStartup": true,
    "solarFallback": "18:00"
  }
}
```

Both user and exported root schedulers honor this setting. Re-enable it and
re-export if a privileged schedule should resume too.

## Keep a phase without scheduling it

Set that phase's `enabled` field to `false`. It remains available for manual
`apply --phase <id>` because manual application selects by ID and does not check
the phase trigger.

## Prevent reapplication after service restart

Set `runtime.reapplyOnStartup` to `false`. ThemeTime will still apply a new phase
or a changed active phase. This is useful when custom commands should run only on
real transitions, but remember that external KDE changes will no longer be
corrected just because the service restarts.

## Change the polar fallback

```json
{
  "runtime": {
    "enabled": true,
    "reapplyOnStartup": true,
    "solarFallback": "07:00"
  }
}
```

The value must be 24-hour `HH:MM`. It is used by both unavailable sunrise and
unavailable sunset, so read the caveat in
[Solar events and offsets](Solar-Events-and-Offsets).

## Test an alternate schedule without replacing the main file

```sh
themetime daemon --once --config /absolute/path/to/experiment.json
```

This uses the alternate JSON for that daemon invocation. Application state and
snapshots still use the normal XDG state directory, so use manual phase apply or
a disposable account when testing destructive custom actions.

The full field-by-field schema is in the
[configuration reference](../docs/configuration.md).
