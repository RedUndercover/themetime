# Actions cookbook

These examples are fragments for a phase's `actions` array. Replace IDs and
absolute paths with values installed on your system. Validate the result with
`themetime show-config`, then test it with `themetime apply --phase <id>`.

Actions layer by setting. You can place a color scheme only in sunrise and
sunset rules while using separate, more granular video rules throughout the day;
the video transitions retain the latest color scheme.

## Light scheme with a custom accent

```json
[
  {
    "type": "colorScheme",
    "value": "BreezeLight",
    "values": { "accent": "#D98743" }
  }
]
```

## Global theme with deliberate overrides

```json
[
  { "type": "globalTheme", "value": "org.kde.breezedark.desktop" },
  { "type": "iconTheme", "value": "Papirus-Dark" },
  { "type": "cursorTheme", "value": "Breeze_Snow", "values": { "size": "24" } },
  { "type": "accentColor", "value": "#7AA2F7", "values": { "colorScheme": "BreezeDark" } }
]
```

ThemeTime applies the global theme before the more specific overrides, even if
the JSON order is different.

## Day and night wallpapers

Day phase:

```json
[
  {
    "type": "staticWallpaper",
    "value": "/home/alex/Wallpapers/solar-day.webp",
    "values": { "fillMode": "2" }
  }
]
```

Night phase:

```json
[
  {
    "type": "staticWallpaper",
    "value": "/home/alex/Wallpapers/observatory-night.webp",
    "values": { "fillMode": "2" }
  }
]
```

ThemeTime accepts a leading `~/` for media and normalizes it on GUI save. An
absolute path is still the clearest form; `$HOME` is not expanded.

## Different wallpaper on a second screen

```json
[
  {
    "type": "staticWallpaper",
    "value": "/home/alex/Wallpapers/portrait-night.png",
    "screen": "1"
  }
]
```

Remove `screen` to target every desktop. Plasma screen identifiers can change
after reconnecting monitors.

## Quiet looping video

```json
[
  {
    "type": "videoWallpaper",
    "value": "/home/alex/Videos/corona-loop.webm",
    "values": {
      "muteMode": "5",
      "volume": "0",
      "batteryPausesVideo": "true",
      "screenOffPausesVideo": "true"
    }
  }
]
```

This requires Smart Video Wallpaper Reborn and `qdbus6`. Plugin option codes are
passed through, so begin with defaults if a plugin update behaves differently.

## Work and reading font profiles

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

Copy serialized font strings from KDE-generated config rather than guessing the
format. Existing applications may need a restart.

## Desktop notification after a phase applies

```json
[
  {
    "type": "customCommand",
    "value": "/usr/bin/notify-send 'ThemeTime' 'Night phase applied'"
  }
]
```

Custom commands execute via `/bin/sh -c` with the service environment. Treat
them as code and use absolute executable paths.

## Login and boot themes

```json
[
  { "type": "sddmTheme", "value": "breeze" },
  { "type": "plymouthTheme", "value": "bgrt" }
]
```

These actions require installed theme IDs and the separate root service. Export
the schedule after saving; see [Privileged themes](Privileged-Themes).

For every optional field and command dependency, consult the full
[action reference](../docs/actions.md).
