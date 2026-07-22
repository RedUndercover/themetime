package scheduler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/RedUndercover/themetime/internal/model"
	"github.com/RedUndercover/themetime/internal/solar"
)

type Transition struct {
	Phase model.Phase
	At    time.Time
}

type Plan struct {
	Transitions      []Transition
	Active           *Transition
	Next             *Transition
	EffectiveActions []EffectiveAction
}

type EffectiveAction struct {
	Key    string
	Action model.Action
	Event  bool
}

type SolarEvent struct {
	Kind model.TriggerKind
	At   time.Time
	Err  error
}

func ResolveDay(cfg model.Config, date time.Time) ([]Transition, error) {
	loc, err := time.LoadLocation(cfg.Location.Timezone)
	if err != nil {
		return nil, err
	}
	day := startOfDay(date.In(loc))
	transitions := make([]Transition, 0, len(cfg.Phases))
	for _, phase := range cfg.Phases {
		if !phase.Enabled {
			continue
		}
		at, err := resolveTrigger(cfg, phase.Start, day, loc)
		if err != nil {
			return nil, fmt.Errorf("phase %q: %w", phase.ID, err)
		}
		transitions = append(transitions, Transition{Phase: phase, At: at})
	}
	sort.SliceStable(transitions, func(i, j int) bool {
		return transitions[i].At.Before(transitions[j].At)
	})
	return transitions, nil
}

func ResolveSolarEvents(cfg model.Config, date time.Time) ([]SolarEvent, error) {
	loc, err := time.LoadLocation(cfg.Location.Timezone)
	if err != nil {
		return nil, err
	}
	day := startOfDay(date.In(loc))
	events := make([]SolarEvent, 0, len(model.SolarTriggerKinds()))
	for _, kind := range model.SolarTriggerKinds() {
		at, err := resolveSolarEvent(cfg, kind, day, loc)
		events = append(events, SolarEvent{Kind: kind, At: at, Err: err})
	}
	return events, nil
}

func ResolveNow(cfg model.Config, now time.Time) (Plan, error) {
	loc, err := time.LoadLocation(cfg.Location.Timezone)
	if err != nil {
		return Plan{}, err
	}
	now = now.In(loc)
	today, err := ResolveDay(cfg, now)
	if err != nil {
		return Plan{}, err
	}
	tomorrow, err := ResolveDay(cfg, now.AddDate(0, 0, 1))
	if err != nil {
		return Plan{}, err
	}
	yesterday, err := ResolveDay(cfg, now.AddDate(0, 0, -1))
	if err != nil {
		return Plan{}, err
	}

	all := append([]Transition{}, yesterday...)
	all = append(all, today...)
	all = append(all, tomorrow...)
	sort.SliceStable(all, func(i, j int) bool {
		return all[i].At.Before(all[j].At)
	})

	plan := Plan{Transitions: today}
	for i := range all {
		if !all[i].At.After(now) {
			t := all[i]
			plan.Active = &t
			continue
		}
		t := all[i]
		plan.Next = &t
		break
	}
	if plan.Active != nil {
		plan.EffectiveActions = composeEffectiveActions(all, *plan.Active, now)
	}
	return plan, nil
}

func (p Plan) EffectivePhase() *model.Phase {
	if p.Active == nil {
		return nil
	}
	phase := p.Active.Phase
	phase.Actions = make([]model.Action, 0, len(p.EffectiveActions))
	for _, effective := range p.EffectiveActions {
		phase.Actions = append(phase.Actions, effective.Action)
	}
	return &phase
}

func PhaseFingerprint(phase model.Phase) string {
	data, _ := json.Marshal(phase)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func ActionFingerprint(action model.Action) string {
	data, _ := json.Marshal(action)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

type effectiveActionSlot struct {
	effective EffectiveAction
	order     int
}

func composeEffectiveActions(transitions []Transition, active Transition, now time.Time) []EffectiveAction {
	slots := map[string]effectiveActionSlot{}
	var events []effectiveActionSlot
	order := 0

	for _, transition := range transitions {
		if transition.At.After(now) {
			break
		}
		actions := append([]model.Action(nil), transition.Phase.Actions...)
		sort.SliceStable(actions, func(i, j int) bool {
			return model.ActionPriority(actions[i].Type) < model.ActionPriority(actions[j].Type)
		})
		isActiveTransition := transition.At.Equal(active.At)
		for _, action := range actions {
			order++
			if action.Type == model.ActionCustomCommand {
				if isActiveTransition {
					events = append(events, effectiveActionSlot{
						effective: EffectiveAction{Action: action, Event: true},
						order:     order,
					})
				}
				continue
			}

			if action.Type == model.ActionGlobalTheme {
				clearAppearanceSlots(slots)
			}
			if action.Type == model.ActionColorScheme {
				delete(slots, actionTrackKey(model.Action{Type: model.ActionAccentColor}))
			}
			key := actionTrackKey(action)
			if isGlobalWallpaper(action) {
				clearWallpaperSlots(slots)
			}
			slots[key] = effectiveActionSlot{
				effective: EffectiveAction{Key: key, Action: action},
				order:     order,
			}
		}
	}

	composed := make([]effectiveActionSlot, 0, len(slots)+len(events))
	for _, slot := range slots {
		composed = append(composed, slot)
	}
	composed = append(composed, events...)
	sort.SliceStable(composed, func(i, j int) bool {
		leftPriority := model.ActionPriority(composed[i].effective.Action.Type)
		rightPriority := model.ActionPriority(composed[j].effective.Action.Type)
		if leftPriority != rightPriority {
			return leftPriority < rightPriority
		}
		return composed[i].order < composed[j].order
	})

	out := make([]EffectiveAction, 0, len(composed))
	for _, slot := range composed {
		out = append(out, slot.effective)
	}
	return out
}

func actionTrackKey(action model.Action) string {
	if action.Type == model.ActionStaticWallpaper || action.Type == model.ActionVideoWallpaper {
		if action.Screen == "" {
			return "wallpaper:all"
		}
		return "wallpaper:screen:" + action.Screen
	}
	return "action:" + string(action.Type)
}

func isGlobalWallpaper(action model.Action) bool {
	return (action.Type == model.ActionStaticWallpaper || action.Type == model.ActionVideoWallpaper) && action.Screen == ""
}

func clearWallpaperSlots(slots map[string]effectiveActionSlot) {
	for key := range slots {
		if strings.HasPrefix(key, "wallpaper:") {
			delete(slots, key)
		}
	}
}

func clearAppearanceSlots(slots map[string]effectiveActionSlot) {
	for _, actionType := range []model.ActionType{
		model.ActionGlobalTheme,
		model.ActionColorScheme,
		model.ActionAccentColor,
		model.ActionPlasmaStyle,
		model.ActionIconTheme,
		model.ActionCursorTheme,
		model.ActionWindowDecoration,
		model.ActionFontProfile,
	} {
		delete(slots, actionTrackKey(model.Action{Type: actionType}))
	}
	clearWallpaperSlots(slots)
}

func resolveTrigger(cfg model.Config, trigger model.Trigger, day time.Time, loc *time.Location) (time.Time, error) {
	switch trigger.Kind {
	case model.TriggerClock:
		return withClock(day, trigger.Clock)
	case model.TriggerAstronomicalDawn, model.TriggerNauticalDawn, model.TriggerCivilDawn, model.TriggerSunrise, model.TriggerSolarNoon, model.TriggerSunset, model.TriggerCivilDusk, model.TriggerNauticalDusk, model.TriggerAstronomicalDusk:
		at, err := resolveSolarEvent(cfg, trigger.Kind, day, loc)
		if err != nil {
			if solarFallback, ok := triggerFallback(day, cfg, trigger.Kind); ok {
				return solarFallback.Add(time.Duration(trigger.OffsetMinutes) * time.Minute), nil
			}
			return time.Time{}, err
		}
		return at.Add(time.Duration(trigger.OffsetMinutes) * time.Minute), nil
	default:
		return time.Time{}, fmt.Errorf("unknown trigger kind %q", trigger.Kind)
	}
}

func resolveSolarEvent(cfg model.Config, kind model.TriggerKind, day time.Time, loc *time.Location) (time.Time, error) {
	switch kind {
	case model.TriggerAstronomicalDawn:
		return solar.AstronomicalDawn(day, cfg.Location.Latitude, cfg.Location.Longitude, loc)
	case model.TriggerNauticalDawn:
		return solar.NauticalDawn(day, cfg.Location.Latitude, cfg.Location.Longitude, loc)
	case model.TriggerCivilDawn:
		return solar.CivilDawn(day, cfg.Location.Latitude, cfg.Location.Longitude, loc)
	case model.TriggerSunrise:
		return solar.Sunrise(day, cfg.Location.Latitude, cfg.Location.Longitude, loc)
	case model.TriggerSolarNoon:
		return solar.SolarNoon(day, cfg.Location.Latitude, cfg.Location.Longitude, loc)
	case model.TriggerSunset:
		return solar.Sunset(day, cfg.Location.Latitude, cfg.Location.Longitude, loc)
	case model.TriggerCivilDusk:
		return solar.CivilDusk(day, cfg.Location.Latitude, cfg.Location.Longitude, loc)
	case model.TriggerNauticalDusk:
		return solar.NauticalDusk(day, cfg.Location.Latitude, cfg.Location.Longitude, loc)
	case model.TriggerAstronomicalDusk:
		return solar.AstronomicalDusk(day, cfg.Location.Latitude, cfg.Location.Longitude, loc)
	default:
		return time.Time{}, fmt.Errorf("unknown solar event %q", kind)
	}
}

func triggerFallback(day time.Time, cfg model.Config, kind model.TriggerKind) (time.Time, bool) {
	if kind == model.TriggerSunrise || kind == model.TriggerSunset {
		return solarFallback(day, cfg.Runtime.SolarFallback)
	}
	switch kind {
	case model.TriggerAstronomicalDawn:
		return clockFallback(day, "05:00")
	case model.TriggerNauticalDawn:
		return clockFallback(day, "05:30")
	case model.TriggerCivilDawn:
		return clockFallback(day, "06:00")
	case model.TriggerSolarNoon:
		return clockFallback(day, "12:00")
	case model.TriggerCivilDusk:
		return clockFallback(day, "18:30")
	case model.TriggerNauticalDusk:
		return clockFallback(day, "19:00")
	case model.TriggerAstronomicalDusk:
		return clockFallback(day, "19:30")
	default:
		return time.Time{}, false
	}
}

func clockFallback(day time.Time, clock string) (time.Time, bool) {
	at, err := withClock(day, clock)
	return at, err == nil
}

func withClock(day time.Time, clock string) (time.Time, error) {
	parts := strings.Split(clock, ":")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid clock %q", clock)
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, err
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, day.Location()), nil
}

func solarFallback(day time.Time, clock string) (time.Time, bool) {
	at, err := withClock(day, clock)
	return at, err == nil
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
