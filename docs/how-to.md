# How-to guide

This guide collects common ThemeTime workflows. For complete field and action
contracts, use the [configuration](configuration.md) and
[action](actions.md) references.

## Build a day and night schedule

Open **Location**, enter decimal coordinates and an IANA timezone such as
`Europe/Berlin`, preview the solar events, and save.

Configure a daytime phase:

1. Select or add the phase.
2. Choose `sunrise` and set an offset such as `+20` minutes.
3. Add a light `colorScheme` and optional daytime wallpaper.
4. Save and use **Apply now**.

Configure an evening phase the same way with `sunset`, an offset such as `-30`
minutes, and a dark color scheme. A negative offset starts the phase before the
event.

Test both phases from a terminal:

```sh
themetime apply --phase morning
themetime apply --phase evening
```

When both succeed, enable automatic scheduling:

```sh
themetime install-user-service
systemctl --user status themetime.service
```

## Choose triggers and offsets

Solar triggers follow outdoor light through the year. Fixed `clock` triggers
follow a human routine at a stable local time.

| Trigger | Typical use |
| --- | --- |
| Astronomical or nautical dawn | Very early, low-light transitions |
| Civil dawn | Pre-sunrise ambient light |
| Sunrise | Day theme or wallpaper |
| Solar noon | Midday adjustment |
| Sunset | Evening theme |
| Civil, nautical, or astronomical dusk | Successively darker night phases |
| Fixed clock | Work, bedtime, or other routine |

Offsets are minutes relative to a solar event:

```text
sunrise +20  → 20 minutes after sunrise
sunset -30   → 30 minutes before sunset
```

ThemeTime evaluates triggers in the configured IANA timezone. Fixed clocks
retain their local wall time across daylight-saving changes. Before the first
trigger of a day, the final phase from the previous day remains active, so a
night phase naturally spans midnight.

At high latitudes, unavailable solar events use configured fallback times.
Sunrise and sunset share `runtime.solarFallback`; use fixed clocks when one
shared value cannot represent both sides of a polar day or night well. See
[Triggers](configuration.md#triggers) for the complete fallback table.

## Layer appearance actions

Actions layer by setting. A phase that changes only the wallpaper retains the
most recently scheduled color scheme, icons, fonts, and other tracks. A global
theme is applied before more specific overrides.

Light colors with a custom accent:

```json
[
  {
    "type": "colorScheme",
    "value": "BreezeLight",
    "values": { "accent": "#D98743" }
  }
]
```

Global theme with deliberate overrides:

```json
[
  { "type": "globalTheme", "value": "org.kde.breezedark.desktop" },
  { "type": "iconTheme", "value": "Papirus-Dark" },
  { "type": "cursorTheme", "value": "Breeze_Snow", "values": { "size": "24" } },
  { "type": "accentColor", "value": "#7AA2F7", "values": { "colorScheme": "BreezeDark" } }
]
```

## Target wallpapers and video

A static wallpaper can target every desktop or a specific screen:

```json
[
  {
    "type": "staticWallpaper",
    "value": "/home/alex/Wallpapers/night.png",
    "screen": "1",
    "values": { "fillMode": "2" }
  }
]
```

Remove `screen` to target every desktop. Screen identifiers can change after a
monitor is reconnected. Media paths may be absolute or begin with `~/`; `$HOME`
is not expanded.

A quiet looping video:

```json
[
  {
    "type": "videoWallpaper",
    "value": "/home/alex/Videos/night.webm",
    "values": {
      "muteMode": "5",
      "volume": "0",
      "batteryPausesVideo": "true",
      "screenOffPausesVideo": "true"
    }
  }
]
```

Video requires Smart Video Wallpaper Reborn and `qdbus6`. Plugin option codes
are passed through, so start with defaults after a plugin upgrade.

## Configure fonts and commands

Use KDE-generated serialized font values rather than guessing their format:

```json
[
  {
    "type": "fontProfile",
    "values": {
      "font": "Noto Sans,11,-1,5,50,0,0,0,0,0",
      "fixed": "Hack,11,-1,5,50,0,0,0,0,0",
      "windowTitleFont": "Noto Sans,11,-1,5,75,0,0,0,0,0"
    }
  }
]
```

Custom commands run through `/bin/sh -c` in the service environment:

```json
[
  {
    "type": "customCommand",
    "value": "/usr/bin/notify-send 'ThemeTime' 'Night phase applied'"
  }
]
```

Treat custom commands as code and use absolute executable paths.

## Edit configuration safely

Stop the daemon while editing JSON by hand:

```sh
systemctl --user stop themetime.service
cp ~/.config/themetime/config.json ~/.config/themetime/config.json.manual-backup
$EDITOR ~/.config/themetime/config.json
themetime show-config
systemctl --user start themetime.service
```

Do not restart the daemon until `show-config` succeeds.

Add a fixed bedtime phase:

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

It remains active past midnight until another phase begins.

## Pause or tune scheduling

Set `runtime.enabled` to `false` to pause both user and exported privileged
scheduling. Re-export the privileged schedule after changing it.

Set a phase's `enabled` field to `false` to exclude it from scheduling while
keeping it available to `themetime apply --phase ID`.

Set `runtime.reapplyOnStartup` to `false` when custom commands should run only
after real transitions or configuration changes. External KDE changes will no
longer be corrected merely because the service restarts.

Change the shared sunrise/sunset fallback with:

```json
{
  "runtime": {
    "enabled": true,
    "reapplyOnStartup": true,
    "solarFallback": "07:00"
  }
}
```

The value must use 24-hour `HH:MM` format.

## Test an alternate schedule

Run one scheduler pass against another file:

```sh
themetime daemon --once --config /absolute/path/to/experiment.json
```

Application state and snapshots still use the normal XDG state directory. Use
manual phase application or a disposable account when testing destructive
custom commands.

## Schedule SDDM or Plymouth themes

Only the restricted root service applies these actions:

```json
[
  { "type": "sddmTheme", "value": "breeze" },
  { "type": "plymouthTheme", "value": "bgrt" }
]
```

Use installed theme IDs, not labels or paths. After every relevant schedule
change, export a new filtered snapshot:

```sh
themetime install-privileged-schedule
```

See [Services and privileged actions](services-and-privileged-actions.md) for
installation, validation, testing, and recovery.
