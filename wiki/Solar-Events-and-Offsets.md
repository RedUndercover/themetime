# Solar events and offsets

## Which event should I use?

| Event | Practical interpretation |
| --- | --- |
| Astronomical dawn | First extremely faint morning light |
| Nautical dawn | Horizon begins to become usable |
| Civil dawn | Outdoor ambient light before sunrise |
| Sunrise | Sun crosses the morning horizon |
| Solar noon | Sun reaches its daily highest point |
| Sunset | Sun crosses the evening horizon |
| Civil dusk | Ordinary outdoor light fades |
| Nautical dusk | Horizon becomes difficult to distinguish |
| Astronomical dusk | Full astronomical night begins |

ThemeTime calculates these events locally from the configured coordinates,
date, and IANA timezone. It does not fetch weather or geocoding data.

## Offset examples

An offset is applied in minutes after calculating the event:

```text
sunrise + 20   → phase begins 20 minutes after sunrise
sunset  - 30   → phase begins 30 minutes before sunset
civilDusk + 0  → phase begins exactly at civil dusk
```

Use offsets to tune the subjective transition. A light scheme may feel right
after sunrise, while a warm accent often works before sunset.

## Fixed time versus solar time

Choose a fixed `clock` trigger when the behavior should follow human routine,
such as a workday start at 09:00. Choose a solar trigger when it should follow
outdoor light.

You can mix both in one schedule:

| Phase | Trigger |
| --- | --- |
| Commute | `07:30` fixed clock |
| Daylight | sunrise +30 |
| Focus | `09:00` fixed clock |
| Evening | sunset -20 |

Remember that the latest transition wins. A later fixed phase can override a
solar phase and remain active until another transition.

## Midnight and the active phase

ThemeTime resolves yesterday, today, and tomorrow together. At 02:00, before
today's first trigger, yesterday's final phase remains active. This is expected
and makes a night phase naturally span midnight.

## Daylight saving time

Solar events and fixed clocks are evaluated in `location.timezone`. Using an
IANA zone such as `America/New_York` lets the timezone database apply seasonal
clock changes. A fixed `07:00` remains 07:00 local wall time; solar events move
according to both daylight and the clock change.

## Polar regions and missing events

At high latitude, some events do not occur on some dates. ThemeTime substitutes
local fallback times, then applies the offset:

```text
astronomical dawn 05:00    civil dusk 18:30
nautical dawn     05:30    nautical dusk 19:00
civil dawn        06:00    astronomical dusk 19:30
solar noon        12:00
```

Sunrise and sunset both use the configured `runtime.solarFallback`, which
defaults to `18:00`. Because one shared value cannot represent both morning and
evening ideally, polar schedules should use fixed clocks for the affected season
or select other fallbacks deliberately.

The full trigger schema and validation rules are in the
[configuration reference](../docs/configuration.md#triggers).

