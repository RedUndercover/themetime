# Your first schedule

This walkthrough creates a light daytime phase and dark evening phase tied to
the real sunrise and sunset at your location.

## Set the observatory location

1. Open ThemeTime and choose **Location**.
2. Enter a label, decimal latitude and longitude, and an IANA timezone.
3. Choose **Preview**.
4. Confirm sunrise, noon, and sunset look reasonable.
5. Save the location.

Example values for Berlin:

```text
Label: Berlin
Latitude: 52.5200
Longitude: 13.4050
Timezone: Europe/Berlin
```

Do not use a timezone abbreviation such as `EST` or `CEST`. An IANA zone handles
daylight-saving changes correctly.

## Configure morning

1. Select the morning phase.
2. Click the sunrise marker on the timeline.
3. Set the offset to `+20` minutes.
4. Choose a light `colorScheme` action.
5. Optionally add a daytime wallpaper.
6. Save.

The phase will follow sunrise throughout the year.

## Configure evening

1. Select the evening phase.
2. Click the sunset marker.
3. Set the offset to `-30` minutes.
4. Choose a dark `colorScheme` action.
5. Optionally add an evening wallpaper.
6. Save.

A negative offset means the change begins before sunset.

## Test without waiting

Use **Apply now** on each phase. Color and wallpaper changes should happen
immediately. Fonts or icons may require applications to restart.

From a terminal you can see per-action results:

```sh
themetime apply --phase morning
themetime apply --phase evening
```

If an action fails, fix it before enabling background scheduling. Start with
`themetime doctor` and the [troubleshooting page](Troubleshooting).

## Enable automatic application

```sh
themetime install-user-service
```

Confirm it is running:

```sh
systemctl --user status themetime.service
```

The daemon checks every 15 seconds by default but does not reapply an unchanged
phase on every check. It applies after a transition, a change to the active
phase's content, or a daemon start when reapply-on-startup is enabled.

## Add more parts of the day

A useful four-phase rhythm is:

| Phase | Trigger | Offset | Typical actions |
| --- | --- | --- | --- |
| Pre-dawn | `nauticalDawn` | `0` | muted dark colors |
| Day | `sunrise` | `+30` | light scheme, bright wallpaper |
| Evening | `sunset` | `-30` | warm accent, dusk wallpaper |
| Night | `astronomicalDusk` | `0` | darkest scheme, reduced visual motion |

Avoid giving phases the same resolved time. Read
[Solar events and offsets](Solar-Events-and-Offsets) before building more complex
schedules.

