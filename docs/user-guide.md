# User guide

## The Solar Observatory

The main screen is organized around a 24-hour timeline. Solar markers show
astronomical, nautical, and civil dawn and dusk, sunrise, solar noon, and sunset
for the configured location and date. Colored phase bands show which phase owns
each part of the day.

The status area identifies the phase active now and the next scheduled
transition. Calculations use the configured location's timezone, not merely the
computer's current timezone.

## Location

Open **Location** to configure:

- a human-readable label;
- latitude from -90 through 90;
- longitude from -180 through 180;
- an IANA timezone such as `Europe/Berlin` or `America/Los_Angeles`.

Use **Preview** before saving. Preview recalculates the timeline without changing
the persisted configuration. Save only after the event times look plausible.

ThemeTime does not geocode a place name or infer coordinates. This keeps solar
calculations deterministic and avoids a network dependency.

## Phases

A phase is a named set of actions with one trigger. Enabled phases compete on a
single daily schedule; the most recent transition at or before the current time
is active.

Each phase contains:

- a unique machine-readable ID;
- a display name and color;
- an enabled switch;
- a fixed-time or solar trigger, with an optional minute offset;
- zero or more actions.

Disable a phase to retain it without scheduling it. Save after changing phase
order, triggers, or actions.

## Set a trigger

For a fixed trigger, choose **Fixed time** and enter a 24-hour `HH:MM` value.

For a solar trigger, click a marker on the timeline or choose the event in the
phase controls. An offset moves the transition relative to the event:

- `+20` means 20 minutes after the event;
- `-30` means 30 minutes before it;
- `0` means exactly at the event.

The interface offers five-minute offset steps from -180 to +180 minutes. The
configuration format stores an integer number of minutes.

## Understand the active phase

ThemeTime resolves transitions for yesterday, today, and tomorrow. This matters
around midnight: if the first transition today is at 07:00, the active phase at
02:00 is normally yesterday's final phase.

When two transitions resolve to the same instant, their stable ordering is based
on the resolved schedule. Avoid identical times if the intended winner matters.

At polar latitudes, some solar events do not occur on a given date. Sunrise and
sunset use `runtime.solarFallback`; other solar events use built-in civil-time
fallbacks. See [Solar events and offsets](../wiki/Solar-Events-and-Offsets.md).

## Add actions

Choose an action type, select or enter its value, and add it to the phase. The
GUI discovers installed KDE themes and common wallpaper/video locations to make
simple actions selectable.

Actions in a phase are applied in a safety-oriented order rather than relying on
their JSON list order:

1. global theme;
2. colors, accent, Plasma style, icons, cursors, decorations, and fonts;
3. wallpapers;
4. custom commands;
5. privileged actions, which are skipped by the user daemon and handled by the
   separate root daemon.

Advanced `values`, per-screen targeting, font profiles, and custom commands are
best configured in JSON. Follow the safe editing procedure in
[Configuration](configuration.md#editing-json-safely).

### Actions layer across rules

A rule does not need to repeat the whole desktop configuration. ThemeTime keeps
the latest scheduled value for each setting. A practical schedule can therefore
set the light theme once at sunrise, switch videos several times through the
day, set a dark theme at sunset, and continue switching evening videos without
losing that dark theme.

Wallpaper actions replace only the matching wallpaper target. An all-screen
wallpaper replaces earlier global and per-screen wallpapers; a later
screen-specific wallpaper overrides just that screen. Custom commands are not
state: they run for their own transition and are not inherited by later rules.

## Save versus apply

**Save** persists the complete location and phase configuration. It does not
necessarily reapply the current phase immediately.

**Apply now** in the rule editor applies that selected rule's explicit actions
on demand. The tray's **Apply current phase** command reconstructs all currently
effective action tracks. The daemon independently applies changed tracks when a
transition or configuration change occurs. With `reapplyOnStartup` enabled, it
reapplies all persistent effective tracks when the daemon starts.

## System tray

Closing the Solar Observatory window hides it rather than stopping the app. The
tray menu provides:

- current and next phase status;
- **Show ThemeTime**;
- **Apply current phase**;
- **Quit**.

Tray status refreshes when the menu opens. The tray process and the background
scheduler are separate: quitting the GUI does not stop an installed systemd user
service.

## System panel

The **System** panel exposes environment checks and user-service installation.
Use it after setup or when a theme action is missing from the action picker.
Running the doctor from the CLI provides the same class of diagnostic detail in
a terminal-friendly form.

## Expected KDE behavior

KDE components do not all reload settings in the same way. Color and wallpaper
changes are usually immediate. Existing applications may keep old icon or font
settings until restarted, and some changes become fully consistent only after a
new Plasma session.
