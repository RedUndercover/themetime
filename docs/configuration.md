# Configuration reference

ThemeTime stores one JSON document at:

```text
$XDG_CONFIG_HOME/themetime/config.json
```

When `XDG_CONFIG_HOME` is unset, the path is
`~/.config/themetime/config.json`. The file is created automatically with mode
`0600` the first time the CLI, GUI, or daemon loads it.

## Complete example

```json
{
  "version": 1,
  "location": {
    "label": "New York",
    "latitude": 40.7128,
    "longitude": -74.006,
    "timezone": "America/New_York",
    "source": "manual"
  },
  "runtime": {
    "enabled": true,
    "reapplyOnStartup": true,
    "solarFallback": "18:00"
  },
  "phases": [
    {
      "id": "morning",
      "name": "Morning",
      "color": "#F2B84B",
      "enabled": true,
      "start": {
        "kind": "sunrise",
        "offsetMinutes": 20
      },
      "actions": [
        {
          "type": "colorScheme",
          "value": "BreezeLight"
        },
        {
          "type": "staticWallpaper",
          "value": "/home/alex/Pictures/morning.jpg",
          "values": {
            "fillMode": "2"
          }
        }
      ]
    },
    {
      "id": "evening",
      "name": "Evening",
      "color": "#5C7AEA",
      "enabled": true,
      "start": {
        "kind": "sunset",
        "offsetMinutes": -30
      },
      "actions": [
        {
          "type": "colorScheme",
          "value": "BreezeDark",
          "values": {
            "accent": "#5C7AEA"
          }
        }
      ]
    }
  ]
}
```

For static and video media, ThemeTime expands a leading `~/` at application time
and the GUI normalizes it to an absolute path on save. Other shell syntax such
as `$HOME`, command substitution, and arbitrary environment variables is not
expanded. Absolute paths remain the most portable and explicit form.

## Top-level fields

| Field | Type | Meaning |
| --- | --- | --- |
| `version` | integer | Schema version. Must be positive; current version is `1`. |
| `location` | object | Coordinates and timezone used for solar calculations. |
| `runtime` | object | Global scheduler behavior. |
| `phases` | array | Ordered collection of scheduled phases. |

If a parsed config has no `version`, ThemeTime treats it as the current schema
version before validation. Unknown JSON fields are ignored by the current Go
decoder, but should not be relied on for forward compatibility.

## `location`

| Field | Type | Validation and use |
| --- | --- | --- |
| `label` | string | Display label; it does not affect calculations. |
| `latitude` | number | Decimal degrees, -90 through 90. North is positive. |
| `longitude` | number | Decimal degrees, -180 through 180. East is positive. |
| `timezone` | string | Required IANA timezone, for example `Asia/Tokyo`. It must be loadable by the host. |
| `source` | string | Informational provenance such as `manual` or `default`. |

Solar events are calculated from latitude and longitude and then expressed in
the configured timezone. A valid but geographically mismatched timezone shifts
the clock display and schedule, so confirm all three values together.

## `runtime`

| Field | Type | Meaning |
| --- | --- | --- |
| `enabled` | boolean | When false, the daemon loads the config but does not apply phases. |
| `reapplyOnStartup` | boolean | Reapply the active user phase when the daemon starts, even when stored state already matches. |
| `solarFallback` | string | `HH:MM` fallback used for unavailable sunrise and sunset events. |

The root daemon receives a filtered copy of the configuration when a privileged
schedule is exported. Changes to the user config are not automatically exported
again.

## `phases[]`

| Field | Type | Meaning |
| --- | --- | --- |
| `id` | string | Required unique identifier used by state and `apply --phase`. |
| `name` | string | Human-readable phase name. |
| `color` | string | Presentation color used by the timeline. |
| `enabled` | boolean | Whether this phase participates in scheduling. |
| `start` | object | One fixed or solar trigger. |
| `actions` | array | Actions applied when the phase becomes active. |

Keep IDs stable after deployment. Renaming an ID makes the scheduler treat it as
a different phase and reapply it. IDs must be non-empty and unique; use short,
portable identifiers such as `pre-dawn` or `work_day` so they remain convenient
with `apply --phase` and state inspection.

## Triggers

A trigger has these fields:

| Field | Type | Meaning |
| --- | --- | --- |
| `kind` | string | One of the trigger identifiers below. |
| `clock` | string | Required `HH:MM` only when `kind` is `clock`. |
| `offsetMinutes` | integer | Minutes before (negative) or after (positive) the base event. |

Supported `kind` values are:

| Identifier | Event |
| --- | --- |
| `clock` | Fixed local wall-clock time |
| `astronomicalDawn` | Morning sun altitude -18° |
| `nauticalDawn` | Morning sun altitude -12° |
| `civilDawn` | Morning sun altitude -6° |
| `sunrise` | Sunrise |
| `solarNoon` | Daily solar noon |
| `sunset` | Sunset |
| `civilDusk` | Evening sun altitude -6° |
| `nauticalDusk` | Evening sun altitude -12° |
| `astronomicalDusk` | Evening sun altitude -18° |

A solar trigger must not include `clock`. Offset minutes are applied after the
base event or its fallback is resolved. The GUI constrains offsets to -180
through +180 in five-minute steps; the JSON validator accepts any integer.

Fixed-time example:

```json
{
  "start": {
    "kind": "clock",
    "clock": "07:30"
  }
}
```

Solar example:

```json
{
  "start": {
    "kind": "civilDusk",
    "offsetMinutes": 15
  }
}
```

## Polar and unavailable-event fallbacks

Solar events can be absent at high latitudes on some dates. ThemeTime uses the
following local clock fallbacks before applying an offset:

| Event | Fallback |
| --- | --- |
| `astronomicalDawn` | `05:00` |
| `nauticalDawn` | `05:30` |
| `civilDawn` | `06:00` |
| `sunrise` | `runtime.solarFallback` |
| `solarNoon` | `12:00` |
| `sunset` | `runtime.solarFallback` |
| `civilDusk` | `18:30` |
| `nauticalDusk` | `19:00` |
| `astronomicalDusk` | `19:30` |

Because sunrise and sunset share `solarFallback`, users in polar regions may
prefer fixed-time phases for seasons where both events are unavailable.

## Actions

Each action has this common shape:

```json
{
  "type": "actionType",
  "value": "primary value",
  "screen": "optional Plasma screen identifier",
  "values": {
    "option": "string value"
  }
}
```

All entries in `values` are strings, including booleans and numbers. Most action
types require either a non-empty `value` or a non-empty `values` object.
`fontProfile` specifically requires `values`; `customCommand` specifically
requires `value`. See [Actions](actions.md) for the complete per-type contract.

## Schedule resolution

The scheduler resolves enabled phases for yesterday, today, and tomorrow in the
configured timezone. The active phase is the latest transition at or before
now; the next phase is the earliest transition after now. Looking across three
days preserves the previous evening phase after midnight.

Persistent actions from those transitions are composed by independent action
track. The newest color action replaces the prior color, for example, while a
new video wallpaper does not remove an earlier icon or color choice. Static and
video wallpapers share a track per target screen. Custom commands are attached
only to their own transition and are not inherited.

The resulting effective phase and each persistent action are fingerprinted. The
daemon applies only changed tracks, reconstructs all persistent tracks when
`reapplyOnStartup` requests it, and records the new state only after every
attempted non-privileged action completes successfully.

## Editing JSON safely

The GUI saves atomically and validates before replacing the config. For manual
editing, use this workflow:

```sh
systemctl --user stop themetime.service
cp ~/.config/themetime/config.json ~/.config/themetime/config.json.manual-backup
$EDITOR ~/.config/themetime/config.json
./bin/themetime show-config
systemctl --user start themetime.service
```

`show-config` parses and structurally validates the file; it exits with an error
instead of printing the config when validation fails. Location preview and the
scheduler additionally verify that the host can load the named timezone. If
your config lives under a custom `XDG_CONFIG_HOME`, adjust the copy and editor
paths accordingly.

Do not edit `state.json` to change the schedule. Deleting state is rarely
necessary; it only causes the active phase to be treated as unapplied on the
next daemon run.

## Snapshots

Before modifying selected KDE configuration files, ThemeTime writes timestamped
copies under:

```text
$XDG_STATE_HOME/themetime/snapshots/
```

The default is `~/.local/state/themetime/snapshots/`. Snapshots are named like
`20260720-143205.kdeglobals.bak` and written with mode `0600`. They are safety
copies, not an automatic rollback mechanism.
