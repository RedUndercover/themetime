package model

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

const CurrentConfigVersion = 1

type Config struct {
	Version  int             `json:"version"`
	Location Location        `json:"location"`
	Runtime  RuntimeSettings `json:"runtime"`
	Phases   []Phase         `json:"phases"`
}

type RuntimeSettings struct {
	Enabled          bool   `json:"enabled"`
	ReapplyOnStartup bool   `json:"reapplyOnStartup"`
	SolarFallback    string `json:"solarFallback"`
}

type Location struct {
	Label     string  `json:"label"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Source    string  `json:"source"`
}

type Phase struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Color   string   `json:"color"`
	Enabled bool     `json:"enabled"`
	Start   Trigger  `json:"start"`
	Actions []Action `json:"actions"`
}

type TriggerKind string

const (
	TriggerClock            TriggerKind = "clock"
	TriggerAstronomicalDawn TriggerKind = "astronomicalDawn"
	TriggerNauticalDawn     TriggerKind = "nauticalDawn"
	TriggerCivilDawn        TriggerKind = "civilDawn"
	TriggerSunrise          TriggerKind = "sunrise"
	TriggerSolarNoon        TriggerKind = "solarNoon"
	TriggerSunset           TriggerKind = "sunset"
	TriggerCivilDusk        TriggerKind = "civilDusk"
	TriggerNauticalDusk     TriggerKind = "nauticalDusk"
	TriggerAstronomicalDusk TriggerKind = "astronomicalDusk"
)

type Trigger struct {
	Kind          TriggerKind `json:"kind"`
	Clock         string      `json:"clock,omitempty"`
	OffsetMinutes int         `json:"offsetMinutes,omitempty"`
}

type ActionType string

const (
	ActionGlobalTheme      ActionType = "globalTheme"
	ActionColorScheme      ActionType = "colorScheme"
	ActionAccentColor      ActionType = "accentColor"
	ActionPlasmaStyle      ActionType = "plasmaStyle"
	ActionIconTheme        ActionType = "iconTheme"
	ActionCursorTheme      ActionType = "cursorTheme"
	ActionWindowDecoration ActionType = "windowDecoration"
	ActionFontProfile      ActionType = "fontProfile"
	ActionStaticWallpaper  ActionType = "staticWallpaper"
	ActionVideoWallpaper   ActionType = "videoWallpaper"
	ActionCustomCommand    ActionType = "customCommand"
	ActionSDDMTheme        ActionType = "sddmTheme"
	ActionPlymouthTheme    ActionType = "plymouthTheme"
)

type Action struct {
	Type   ActionType        `json:"type"`
	Value  string            `json:"value,omitempty"`
	Screen string            `json:"screen,omitempty"`
	Values map[string]string `json:"values,omitempty"`
}

var clockPattern = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

func DefaultConfig() Config {
	return Config{
		Version: CurrentConfigVersion,
		Location: Location{
			Label:     "New York",
			Latitude:  40.7128,
			Longitude: -74.0060,
			Timezone:  "America/New_York",
			Source:    "default",
		},
		Runtime: RuntimeSettings{
			Enabled:          true,
			ReapplyOnStartup: true,
			SolarFallback:    "18:00",
		},
		Phases: []Phase{
			{
				ID:      "morning",
				Name:    "Morning",
				Color:   "#F2B84B",
				Enabled: true,
				Start: Trigger{
					Kind:          TriggerSunrise,
					OffsetMinutes: 20,
				},
				Actions: []Action{
					{Type: ActionColorScheme, Value: "BreezeLight"},
				},
			},
			{
				ID:      "evening",
				Name:    "Evening",
				Color:   "#5C7AEA",
				Enabled: true,
				Start: Trigger{
					Kind:          TriggerSunset,
					OffsetMinutes: -30,
				},
				Actions: []Action{
					{Type: ActionColorScheme, Value: "BreezeDark"},
				},
			},
		},
	}
}

func (c Config) Validate() error {
	var errs []error
	if c.Version <= 0 {
		errs = append(errs, errors.New("version must be positive"))
	}
	if c.Location.Latitude < -90 || c.Location.Latitude > 90 {
		errs = append(errs, fmt.Errorf("latitude %.4f is outside -90..90", c.Location.Latitude))
	}
	if c.Location.Longitude < -180 || c.Location.Longitude > 180 {
		errs = append(errs, fmt.Errorf("longitude %.4f is outside -180..180", c.Location.Longitude))
	}
	if strings.TrimSpace(c.Location.Timezone) == "" {
		errs = append(errs, errors.New("location timezone is required"))
	}
	if !clockPattern.MatchString(c.Runtime.SolarFallback) {
		errs = append(errs, fmt.Errorf("solar fallback %q must be HH:MM", c.Runtime.SolarFallback))
	}
	ids := map[string]bool{}
	for i, phase := range c.Phases {
		if strings.TrimSpace(phase.ID) == "" {
			errs = append(errs, fmt.Errorf("phase %d id is required", i+1))
		}
		if ids[phase.ID] {
			errs = append(errs, fmt.Errorf("phase id %q is duplicated", phase.ID))
		}
		ids[phase.ID] = true
		if err := phase.Start.Validate(); err != nil {
			errs = append(errs, fmt.Errorf("phase %q trigger: %w", phase.ID, err))
		}
		for _, action := range phase.Actions {
			if err := action.Validate(); err != nil {
				errs = append(errs, fmt.Errorf("phase %q action %q: %w", phase.ID, action.Type, err))
			}
		}
	}
	return errors.Join(errs...)
}

func (t Trigger) Validate() error {
	switch t.Kind {
	case TriggerClock:
		if !clockPattern.MatchString(t.Clock) {
			return fmt.Errorf("clock trigger %q must be HH:MM", t.Clock)
		}
	case TriggerAstronomicalDawn, TriggerNauticalDawn, TriggerCivilDawn, TriggerSunrise, TriggerSolarNoon, TriggerSunset, TriggerCivilDusk, TriggerNauticalDusk, TriggerAstronomicalDusk:
		if t.Clock != "" {
			return errors.New("solar trigger must not include clock")
		}
	default:
		return fmt.Errorf("unknown trigger kind %q", t.Kind)
	}
	return nil
}

func SolarTriggerKinds() []TriggerKind {
	return []TriggerKind{
		TriggerAstronomicalDawn,
		TriggerNauticalDawn,
		TriggerCivilDawn,
		TriggerSunrise,
		TriggerSolarNoon,
		TriggerSunset,
		TriggerCivilDusk,
		TriggerNauticalDusk,
		TriggerAstronomicalDusk,
	}
}

func IsSolarTrigger(kind TriggerKind) bool {
	return slices.Contains(SolarTriggerKinds(), kind)
}

func SolarTriggerLabel(kind TriggerKind) string {
	switch kind {
	case TriggerAstronomicalDawn:
		return "Astronomical dawn"
	case TriggerNauticalDawn:
		return "Nautical dawn"
	case TriggerCivilDawn:
		return "Civil dawn"
	case TriggerSunrise:
		return "Sunrise"
	case TriggerSolarNoon:
		return "Solar noon"
	case TriggerSunset:
		return "Sunset"
	case TriggerCivilDusk:
		return "Civil dusk"
	case TriggerNauticalDusk:
		return "Nautical dusk"
	case TriggerAstronomicalDusk:
		return "Astronomical dusk"
	default:
		return string(kind)
	}
}

func SolarTriggerShortLabel(kind TriggerKind) string {
	switch kind {
	case TriggerAstronomicalDawn:
		return "Astro dawn"
	case TriggerNauticalDawn:
		return "Nautical dawn"
	case TriggerCivilDawn:
		return "Civil dawn"
	case TriggerSunrise:
		return "Sunrise"
	case TriggerSolarNoon:
		return "Noon"
	case TriggerSunset:
		return "Sunset"
	case TriggerCivilDusk:
		return "Civil dusk"
	case TriggerNauticalDusk:
		return "Nautical dusk"
	case TriggerAstronomicalDusk:
		return "Astro dusk"
	default:
		return string(kind)
	}
}

func (a Action) Validate() error {
	valid := slices.Contains(AllActionTypes(), a.Type)
	if !valid {
		return fmt.Errorf("unknown action type %q", a.Type)
	}
	switch a.Type {
	case ActionFontProfile:
		if len(a.Values) == 0 {
			return errors.New("font profile requires values")
		}
	case ActionCustomCommand:
		if strings.TrimSpace(a.Value) == "" {
			return errors.New("custom command requires a command string")
		}
	default:
		if strings.TrimSpace(a.Value) == "" && len(a.Values) == 0 {
			return errors.New("action requires a value")
		}
	}
	return nil
}

func AllActionTypes() []ActionType {
	return []ActionType{
		ActionGlobalTheme,
		ActionColorScheme,
		ActionAccentColor,
		ActionPlasmaStyle,
		ActionIconTheme,
		ActionCursorTheme,
		ActionWindowDecoration,
		ActionFontProfile,
		ActionStaticWallpaper,
		ActionVideoWallpaper,
		ActionCustomCommand,
		ActionSDDMTheme,
		ActionPlymouthTheme,
	}
}

// ActionPriority defines the application order for actions that share a phase
// or are composed from multiple phases. Lower values run first.
func ActionPriority(actionType ActionType) int {
	switch actionType {
	case ActionGlobalTheme:
		return 0
	case ActionColorScheme, ActionAccentColor, ActionPlasmaStyle, ActionIconTheme, ActionCursorTheme, ActionWindowDecoration, ActionFontProfile:
		return 1
	case ActionStaticWallpaper, ActionVideoWallpaper:
		return 2
	case ActionCustomCommand:
		return 3
	case ActionSDDMTheme, ActionPlymouthTheme:
		return 4
	default:
		return 5
	}
}

func (a Action) IsPrivileged() bool {
	return a.Type == ActionSDDMTheme || a.Type == ActionPlymouthTheme
}

func (a Action) Label() string {
	switch a.Type {
	case ActionGlobalTheme:
		return "Global theme"
	case ActionColorScheme:
		return "Color scheme"
	case ActionAccentColor:
		return "Accent color"
	case ActionPlasmaStyle:
		return "Plasma style"
	case ActionIconTheme:
		return "Icon theme"
	case ActionCursorTheme:
		return "Cursor theme"
	case ActionWindowDecoration:
		return "Window decoration"
	case ActionFontProfile:
		return "Fonts"
	case ActionStaticWallpaper:
		return "Static wallpaper"
	case ActionVideoWallpaper:
		return "Video wallpaper"
	case ActionCustomCommand:
		return "Command"
	case ActionSDDMTheme:
		return "SDDM theme"
	case ActionPlymouthTheme:
		return "Plymouth theme"
	default:
		return string(a.Type)
	}
}

func (t Trigger) String() string {
	switch t.Kind {
	case TriggerClock:
		return t.Clock
	case TriggerAstronomicalDawn, TriggerNauticalDawn, TriggerCivilDawn, TriggerSunrise, TriggerSolarNoon, TriggerSunset, TriggerCivilDusk, TriggerNauticalDusk, TriggerAstronomicalDusk:
		base := SolarTriggerLabel(t.Kind)
		if t.OffsetMinutes == 0 {
			return base
		}
		sign := "+"
		offset := t.OffsetMinutes
		if offset < 0 {
			sign = "-"
			offset = -offset
		}
		return fmt.Sprintf("%s %s %dm", base, sign, offset)
	default:
		return string(t.Kind)
	}
}
