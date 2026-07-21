# Action reference

Actions describe what ThemeTime changes when a phase becomes active. Actions are
layered across rules: omitting a setting does not clear it. The user daemon
applies desktop-session actions; the root daemon separately handles only
`sddmTheme` and `plymouthTheme`.

## Independent action tracks

ThemeTime resolves the latest scheduled value for each persistent action type.
For example, if sunrise sets a color scheme and later rules change only videos,
the sunrise color scheme remains effective. The daemon records fingerprints per
track and applies only tracks whose desired value changed.

Most action types have one global track, including colors, accents, Plasma
style, icons, cursors, decorations, and fonts. Static and video wallpaper types
share wallpaper tracks because they target the same Plasma setting:

- an action with an empty `screen` replaces all earlier wallpaper tracks;
- a screen-specific action replaces only that screen's wallpaper track and can
  layer over an earlier all-screen wallpaper;
- static and video actions targeting the same screen replace each other.

A later global theme establishes a new appearance baseline and clears inherited
color, accent, Plasma-style, icon, cursor, decoration, font, and wallpaper
tracks. More specific actions at the same or later transition layer over that
baseline. A later color scheme clears an older standalone accent; an accent at
the same or later transition overrides the scheme accent.

`customCommand` is event-only. It runs when its own transition becomes active
but is not inherited by later rules or replayed merely because persistent tracks
are reconstructed.

## Summary

| Type | Primary `value` | Additional `values` | Required command or component |
| --- | --- | --- | --- |
| `globalTheme` | Look-and-feel package ID | — | `plasma-apply-lookandfeel` |
| `colorScheme` | Color scheme ID | `accent` | `plasma-apply-colorscheme` |
| `accentColor` | Color value | `colorScheme` | `plasma-apply-colorscheme`, `kreadconfig6` |
| `plasmaStyle` | Desktop theme ID | — | `plasma-apply-desktoptheme` |
| `iconTheme` | Icon theme ID | — | `kwriteconfig6`; optional `kbuildsycoca6` |
| `cursorTheme` | Cursor theme ID | `size` | `plasma-apply-cursortheme` |
| `windowDecoration` | Decoration theme | `library`, `theme` | `kwriteconfig6`; optional `qdbus6` |
| `fontProfile` | Usually omitted | Font keys | `kwriteconfig6`; optional `qdbus6` |
| `staticWallpaper` | Image path or URI | `fillMode` | `plasma-apply-wallpaperimage` or `qdbus6` |
| `videoWallpaper` | Video path | Playback/plugin options | Smart Video Wallpaper Reborn, `qdbus6` |
| `customCommand` | Shell command | — | `/bin/sh` |
| `sddmTheme` | Installed SDDM theme ID | — | Privileged service |
| `plymouthTheme` | Installed Plymouth theme ID | — | Privileged service, `plymouth-set-default-theme` |

The GUI focuses on installed themes and simple values it can discover. Advanced
maps, screen targeting, font profiles, and custom commands can be authored in
JSON.

## Ordering and failures

ThemeTime groups actions into this order:

1. `globalTheme`;
2. colors, accent, Plasma style, icons, cursors, decorations, and fonts;
3. static and video wallpapers;
4. `customCommand`;
5. privileged actions.

Do not use list order to express dependencies between actions in the same group;
their relative ordering is an implementation detail. A global theme comes first
because it may reset several settings that later actions intentionally override.

The manual `apply --phase` command applies only actions explicitly attached to
the selected phase and reports one result per action. The tray's **Apply current
phase** operation reconstructs the complete effective state. In the daemon, any
failed desktop action prevents the new state from being recorded, so changed
tracks are attempted again on a later poll. Privileged actions are reported as
skipped by the user applier and are independently deduplicated by the root
daemon.

## `globalTheme`

Applies a KDE look-and-feel package:

```json
{ "type": "globalTheme", "value": "org.kde.breeze.desktop" }
```

Equivalent operation:

```text
plasma-apply-lookandfeel --apply <value>
```

Package IDs are not necessarily the same as display names.

## `colorScheme`

Applies a KDE color scheme, optionally overriding its accent:

```json
{
  "type": "colorScheme",
  "value": "BreezeDark",
  "values": { "accent": "#6EA8FE" }
}
```

The value is passed to `plasma-apply-colorscheme`. Installed scheme IDs can
differ from the label displayed by System Settings.

## `accentColor`

Applies an accent while retaining the current color scheme:

```json
{ "type": "accentColor", "value": "#F2B84B" }
```

ThemeTime reads `General/ColorScheme` from `kdeglobals`. To make the result
independent of the current state, provide a scheme explicitly:

```json
{
  "type": "accentColor",
  "value": "#F2B84B",
  "values": { "colorScheme": "BreezeLight" }
}
```

The action fails if no scheme is supplied and the current scheme cannot be
read.

## `plasmaStyle`

```json
{ "type": "plasmaStyle", "value": "breeze-dark" }
```

This calls `plasma-apply-desktoptheme` with the desktop theme package ID.

## `iconTheme`

```json
{ "type": "iconTheme", "value": "breeze-dark" }
```

ThemeTime snapshots `kdeglobals`, writes `Icons/Theme`, and runs
`kbuildsycoca6 --noincremental` when available. Existing applications may not
refresh all icons until restarted.

## `cursorTheme`

```json
{
  "type": "cursorTheme",
  "value": "Breeze_Snow",
  "values": { "size": "24" }
}
```

`size` is optional and is passed through to
`plasma-apply-cursortheme --size`.

## `windowDecoration`

The short form uses the Breeze decoration library:

```json
{ "type": "windowDecoration", "value": "Breeze" }
```

The expanded form controls both KWin values:

```json
{
  "type": "windowDecoration",
  "values": {
    "library": "org.kde.breeze",
    "theme": "Breeze"
  }
}
```

If `value` is non-empty it takes precedence over `values.theme`. ThemeTime
snapshots `kwinrc` and asks KWin to reconfigure when `qdbus6` is available.

## `fontProfile`

A font profile writes one or more KDE font strings into the `General` group of
`kdeglobals`:

```json
{
  "type": "fontProfile",
  "values": {
    "font": "Noto Sans,10,-1,5,50,0,0,0,0,0",
    "fixed": "Hack,10,-1,5,50,0,0,0,0,0",
    "smallestFont": "Noto Sans,8,-1,5,50,0,0,0,0,0",
    "windowTitleFont": "Noto Sans,10,-1,5,75,0,0,0,0,0"
  }
}
```

Recognized input keys are:

```text
font
fixed
smallestFont
toolBarFont
menuFont
activeFont
desktopFont
taskbarFont
windowTitle
windowTitleFont
```

`windowTitle` and `windowTitleFont` both write KDE's `activeFont` key. Empty
entries are skipped. ThemeTime does not parse or normalize KDE's serialized font
format; use a value produced by KDE configuration when possible.

## `staticWallpaper`

```json
{
  "type": "staticWallpaper",
  "value": "/home/alex/Pictures/observatory-dawn.jpg",
  "values": { "fillMode": "2" }
}
```

With no `screen`, ThemeTime prefers `plasma-apply-wallpaperimage` when installed.
Otherwise it configures Plasma's `org.kde.image` wallpaper plugin through a
Plasma shell script. `fillMode` defaults to `2`; it is passed through as a
plugin/command value.

To target one Plasma screen, include the numeric screen identifier as a string:

```json
{
  "type": "staticWallpaper",
  "value": "/home/alex/Pictures/portrait.png",
  "screen": "1"
}
```

An empty `screen` applies the scripted action to every desktop. Screen IDs are
Plasma identifiers and can change when display topology changes.

The GUI scans up to three levels beneath `~/Pictures`, `~/Wallpapers`,
`/usr/share/wallpapers`, and `/usr/share/backgrounds`, returning at most 250
supported images. Recognized extensions are AVIF, BMP, JPEG/JPG, PNG, SVG, and
WebP.

## `videoWallpaper`

```json
{
  "type": "videoWallpaper",
  "value": "/home/alex/Videos/solar-loop.webm",
  "values": {
    "fillMode": "2",
    "muteMode": "5",
    "volume": "0",
    "batteryPausesVideo": "true"
  }
}
```

This targets Smart Video Wallpaper Reborn with plugin ID
`luisbocanegra.smart.video.wallpaper.reborn`. The action requires `qdbus6` and
the plugin installed for the current user or system.

Supported optional keys are:

| Key | Default |
| --- | --- |
| `duration` | Plugin/video duration (`0`) |
| `muteMode` | `5` |
| `fillMode` | `2` |
| `pauseMode` | `3` |
| `batteryPausesVideo` | `true` |
| `screenOffPausesVideo` | `true` |
| `crossfadeEnabled` | `false` |
| `checkWindowsActiveScreen` | `true` |
| `alternativePlaybackRateMode` | `3` |
| `volume` | Not written |
| `blurMode` | Not written |

The generated video item loops, starts enabled, uses playback rate 1, and
resumes the last video. A positive `duration` marks the duration as custom.
Plugin option codes are deliberately passed through; consult the installed
plugin version when changing them.

The GUI searches `~/Videos` and `~/Wallpapers` to depth three, up to 250 files,
and recognizes AVI, M4V, MKV, MOV, MP4, and WebM. `screen` works the same way as
for static wallpapers.

## `customCommand`

```json
{
  "type": "customCommand",
  "value": "notify-send 'ThemeTime' 'Evening phase applied'"
}
```

ThemeTime executes the value as:

```text
/bin/sh -c <value>
```

This is intentionally powerful and should be treated as code. Never paste an
untrusted schedule or command into the config. Commands launched by the systemd
user service inherit its service environment, which may have a different `PATH`
from an interactive shell; use absolute executable paths when reliability
matters.

Custom commands are never accepted by the privileged helper.

## `sddmTheme`

```json
{ "type": "sddmTheme", "value": "breeze" }
```

The root daemon verifies that the theme exists under `/usr/share/sddm/themes` or
`/usr/local/share/sddm/themes`, snapshots the previous ThemeTime SDDM fragment,
and writes `/etc/sddm.conf.d/90-themetime.conf`.

Saving the user config is not enough: export the filtered privileged schedule
after every relevant change. See
[Services and privileged actions](services-and-privileged-actions.md).

## `plymouthTheme`

```json
{ "type": "plymouthTheme", "value": "bgrt" }
```

The root daemon verifies the name using `plymouth-set-default-theme --list`, then
runs:

```text
plymouth-set-default-theme -R <value>
```

`-R` rebuilds the boot image and can be expensive. Root-side state and phase
fingerprints prevent repeating it every poll.
