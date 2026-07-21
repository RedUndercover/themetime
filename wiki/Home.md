# ThemeTime wiki

ThemeTime changes KDE Plasma 6 themes and wallpapers on a solar or clock-based
schedule. The Solar Observatory UI turns the day into a timeline: set your
location, anchor phases to sunrise or twilight, add appearance actions, and let
the background service apply them.

## Choose a path

New installation:

1. [Install ThemeTime](Installation).
2. [Build your first schedule](Your-First-Schedule).
3. [Enable the background service](System-Tray-and-Background-Service).

Customize an existing setup:

- Learn [solar events and offsets](Solar-Events-and-Offsets).
- Copy patterns from the [actions cookbook](Actions-Cookbook).
- Use [configuration recipes](Configuration-Recipes) for advanced JSON features.
- Schedule the login screen or boot splash with
  [privileged themes](Privileged-Themes).

Something is not working:

- Follow the [wiki troubleshooting checklist](Troubleshooting).
- Then use the full [troubleshooting reference](../docs/troubleshooting.md).

Contributing:

- Start with the [developer guide](Developer-Guide).
- Use the full [architecture reference](../docs/architecture.md) for component
  boundaries and failure behavior.

## Mental model

A schedule contains phases. Each phase has one start trigger and any number of
actions. The phase whose transition most recently occurred is active:

```text
astronomical dawn   sunrise                 sunset       astronomical dusk
       │               │                      │                 │
───────┼── pre-dawn ───┼── morning/day ──────┼── evening ──────┼── night ──
```

Solar triggers move through the year because ThemeTime calculates them for the
configured coordinates and date. Fixed triggers stay at the same local clock
time. Offsets let a phase begin before or after an event.

When a phase becomes active, its actions layer onto values established by earlier
phases. This makes granular wallpaper/video phases independent of day/night
themes. The normal daemon runs as your user. An optional root daemon separately
handles only SDDM and Plymouth theme IDs.

## What ThemeTime can change

- KDE global themes, color schemes, and accents;
- Plasma styles, icon and cursor themes;
- window decorations and font profiles;
- static and Smart Video Wallpaper Reborn wallpapers;
- user shell commands;
- SDDM login and Plymouth boot themes through the restricted root service.

See the [action reference](../docs/actions.md) for required commands and every
advanced option.

## Project documentation

The wiki favors short workflows. The versioned docs directory is the canonical
technical reference:

- [Getting started](../docs/getting-started.md)
- [User guide](../docs/user-guide.md)
- [Configuration reference](../docs/configuration.md)
- [CLI reference](../docs/cli.md)
- [Services and security](../docs/services-and-privileged-actions.md)
- [Packaging and releases](../docs/packaging-and-releases.md)

The Markdown files in this directory can be used directly in the repository or
pushed to a GitHub wiki repository. `Home.md` is the wiki landing page and
`_Sidebar.md` supplies navigation.
