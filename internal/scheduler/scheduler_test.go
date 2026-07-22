package scheduler

import (
	"testing"
	"time"

	"github.com/RedUndercover/themetime/internal/model"
)

func TestResolveNowOvernight(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Phases = []model.Phase{
		{ID: "day", Name: "Day", Enabled: true, Start: model.Trigger{Kind: model.TriggerClock, Clock: "08:00"}, Actions: []model.Action{{Type: model.ActionColorScheme, Value: "BreezeLight"}}},
		{ID: "night", Name: "Night", Enabled: true, Start: model.Trigger{Kind: model.TriggerClock, Clock: "20:00"}, Actions: []model.Action{{Type: model.ActionColorScheme, Value: "BreezeDark"}}},
	}
	loc := mustLocation(t, cfg.Location.Timezone)
	plan, err := ResolveNow(cfg, time.Date(2026, 7, 7, 2, 30, 0, 0, loc))
	if err != nil {
		t.Fatal(err)
	}
	if plan.Active == nil || plan.Active.Phase.ID != "night" {
		t.Fatalf("active phase = %#v, want night", plan.Active)
	}
	if plan.Next == nil || plan.Next.Phase.ID != "day" {
		t.Fatalf("next phase = %#v, want day", plan.Next)
	}
}

func TestResolveSolarOffset(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Phases = []model.Phase{
		{ID: "sunset", Name: "Sunset", Enabled: true, Start: model.Trigger{Kind: model.TriggerSunset, OffsetMinutes: -30}, Actions: []model.Action{{Type: model.ActionColorScheme, Value: "BreezeDark"}}},
	}
	loc := mustLocation(t, cfg.Location.Timezone)
	transitions, err := ResolveDay(cfg, time.Date(2026, 7, 7, 12, 0, 0, 0, loc))
	if err != nil {
		t.Fatal(err)
	}
	if len(transitions) != 1 {
		t.Fatalf("len transitions = %d, want 1", len(transitions))
	}
	got := transitions[0].At
	if got.Hour() < 19 || got.Hour() > 21 {
		t.Fatalf("sunset offset resolved to %v, expected evening time", got)
	}
}

func TestSolarEventsAreDateSpecific(t *testing.T) {
	cfg := model.DefaultConfig()
	loc := mustLocation(t, cfg.Location.Timezone)
	summer := sunsetForDate(t, cfg, time.Date(2026, 7, 7, 12, 0, 0, 0, loc))
	winter := sunsetForDate(t, cfg, time.Date(2026, 12, 7, 12, 0, 0, 0, loc))
	if summer.Equal(winter) {
		t.Fatalf("sunset should vary by date, got %s for both days", summer.Format(time.RFC3339))
	}
	if minuteOfDayForTest(summer) <= minuteOfDayForTest(winter) {
		t.Fatalf("summer sunset = %s, winter sunset = %s; expected summer to be later in New York", summer.Format("15:04"), winter.Format("15:04"))
	}
}

func TestResolveTwilightEvents(t *testing.T) {
	cfg := model.DefaultConfig()
	loc := mustLocation(t, cfg.Location.Timezone)
	events, err := ResolveSolarEvents(cfg, time.Date(2026, 7, 7, 12, 0, 0, 0, loc))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != len(model.SolarTriggerKinds()) {
		t.Fatalf("events = %d, want %d", len(events), len(model.SolarTriggerKinds()))
	}
	seen := map[model.TriggerKind]time.Time{}
	for _, event := range events {
		if event.Err != nil {
			t.Fatalf("%s failed: %v", event.Kind, event.Err)
		}
		seen[event.Kind] = event.At
	}
	if !seen[model.TriggerAstronomicalDawn].Before(seen[model.TriggerNauticalDawn]) {
		t.Fatalf("astronomical dawn should be before nautical dawn")
	}
	if !seen[model.TriggerNauticalDawn].Before(seen[model.TriggerCivilDawn]) {
		t.Fatalf("nautical dawn should be before civil dawn")
	}
	if !seen[model.TriggerCivilDawn].Before(seen[model.TriggerSunrise]) {
		t.Fatalf("civil dawn should be before sunrise")
	}
	if !seen[model.TriggerSunset].Before(seen[model.TriggerCivilDusk]) {
		t.Fatalf("sunset should be before civil dusk")
	}
	if !seen[model.TriggerCivilDusk].Before(seen[model.TriggerNauticalDusk]) {
		t.Fatalf("civil dusk should be before nautical dusk")
	}
	if !seen[model.TriggerNauticalDusk].Before(seen[model.TriggerAstronomicalDusk]) {
		t.Fatalf("nautical dusk should be before astronomical dusk")
	}
}

func TestResolveCivilDuskOffset(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Phases = []model.Phase{
		{ID: "civil-dusk", Name: "Civil dusk", Enabled: true, Start: model.Trigger{Kind: model.TriggerCivilDusk, OffsetMinutes: 10}, Actions: []model.Action{{Type: model.ActionColorScheme, Value: "BreezeDark"}}},
	}
	loc := mustLocation(t, cfg.Location.Timezone)
	transitions, err := ResolveDay(cfg, time.Date(2026, 7, 7, 12, 0, 0, 0, loc))
	if err != nil {
		t.Fatal(err)
	}
	got := transitions[0].At
	if got.Hour() < 20 || got.Hour() > 22 {
		t.Fatalf("civil dusk offset resolved to %v, expected evening twilight", got)
	}
}

func TestResolveSolarFallbackForPolarDate(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Location = model.Location{Label: "Barrow", Latitude: 71.2906, Longitude: -156.7886, Timezone: "America/Anchorage"}
	cfg.Runtime.SolarFallback = "18:00"
	cfg.Phases = []model.Phase{
		{ID: "night", Name: "Night", Enabled: true, Start: model.Trigger{Kind: model.TriggerSunset, OffsetMinutes: 15}, Actions: []model.Action{{Type: model.ActionColorScheme, Value: "BreezeDark"}}},
	}
	loc := mustLocation(t, cfg.Location.Timezone)
	transitions, err := ResolveDay(cfg, time.Date(2026, 6, 20, 12, 0, 0, 0, loc))
	if err != nil {
		t.Fatal(err)
	}
	got := transitions[0].At
	if got.Hour() != 18 || got.Minute() != 15 {
		t.Fatalf("fallback = %s, want 18:15", got.Format("15:04"))
	}
}

func TestResolveNowComposesIndependentActionTracks(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Phases = []model.Phase{
		{
			ID: "theme", Name: "Day theme", Enabled: true,
			Start: model.Trigger{Kind: model.TriggerClock, Clock: "06:00"},
			Actions: []model.Action{
				{Type: model.ActionColorScheme, Value: "BreezeLight"},
				{Type: model.ActionIconTheme, Value: "breeze"},
			},
		},
		{
			ID: "dawn-video", Name: "Dawn video", Enabled: true,
			Start:   model.Trigger{Kind: model.TriggerClock, Clock: "07:00"},
			Actions: []model.Action{{Type: model.ActionVideoWallpaper, Value: "/videos/dawn.mp4"}},
		},
		{
			ID: "day-video", Name: "Day video", Enabled: true,
			Start:   model.Trigger{Kind: model.TriggerClock, Clock: "08:00"},
			Actions: []model.Action{{Type: model.ActionVideoWallpaper, Value: "/videos/day.mp4"}},
		},
	}
	loc := mustLocation(t, cfg.Location.Timezone)
	plan, err := ResolveNow(cfg, time.Date(2026, 7, 7, 9, 0, 0, 0, loc))
	if err != nil {
		t.Fatal(err)
	}
	if plan.Active == nil || plan.Active.Phase.ID != "day-video" {
		t.Fatalf("active = %#v, want day-video", plan.Active)
	}

	got := effectiveActionsByKey(plan.EffectiveActions)
	assertEffectiveAction(t, got, "action:colorScheme", "BreezeLight")
	assertEffectiveAction(t, got, "action:iconTheme", "breeze")
	assertEffectiveAction(t, got, "wallpaper:all", "/videos/day.mp4")
	if len(got) != 3 {
		t.Fatalf("effective actions = %#v, want three independent tracks", got)
	}
}

func TestResolveNowLayersPerScreenWallpaperOverGlobalWallpaper(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Phases = []model.Phase{
		{
			ID: "global", Name: "Global wallpaper", Enabled: true,
			Start:   model.Trigger{Kind: model.TriggerClock, Clock: "06:00"},
			Actions: []model.Action{{Type: model.ActionStaticWallpaper, Value: "/images/all.jpg"}},
		},
		{
			ID: "screen", Name: "Screen override", Enabled: true,
			Start:   model.Trigger{Kind: model.TriggerClock, Clock: "07:00"},
			Actions: []model.Action{{Type: model.ActionVideoWallpaper, Value: "/videos/one.mp4", Screen: "1"}},
		},
	}
	loc := mustLocation(t, cfg.Location.Timezone)
	plan, err := ResolveNow(cfg, time.Date(2026, 7, 7, 8, 0, 0, 0, loc))
	if err != nil {
		t.Fatal(err)
	}
	got := effectiveActionsByKey(plan.EffectiveActions)
	assertEffectiveAction(t, got, "wallpaper:all", "/images/all.jpg")
	assertEffectiveAction(t, got, "wallpaper:screen:1", "/videos/one.mp4")
}

func TestResolveNowDoesNotInheritCustomCommands(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Phases = []model.Phase{
		{
			ID: "command", Name: "Command", Enabled: true,
			Start:   model.Trigger{Kind: model.TriggerClock, Clock: "06:00"},
			Actions: []model.Action{{Type: model.ActionCustomCommand, Value: "notify-send dawn"}},
		},
		{
			ID: "video", Name: "Video", Enabled: true,
			Start:   model.Trigger{Kind: model.TriggerClock, Clock: "07:00"},
			Actions: []model.Action{{Type: model.ActionVideoWallpaper, Value: "/videos/day.mp4"}},
		},
	}
	loc := mustLocation(t, cfg.Location.Timezone)
	plan, err := ResolveNow(cfg, time.Date(2026, 7, 7, 8, 0, 0, 0, loc))
	if err != nil {
		t.Fatal(err)
	}
	for _, effective := range plan.EffectiveActions {
		if effective.Action.Type == model.ActionCustomCommand {
			t.Fatalf("inherited custom command = %#v", effective)
		}
	}
}

func TestResolveNowIncludesActionsFromRulesSharingActiveTransition(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Phases = []model.Phase{
		{
			ID: "theme", Name: "Theme", Enabled: true,
			Start:   model.Trigger{Kind: model.TriggerClock, Clock: "06:00"},
			Actions: []model.Action{{Type: model.ActionColorScheme, Value: "BreezeLight"}},
		},
		{
			ID: "video", Name: "Video", Enabled: true,
			Start: model.Trigger{Kind: model.TriggerClock, Clock: "06:00"},
			Actions: []model.Action{
				{Type: model.ActionVideoWallpaper, Value: "/videos/day.mp4"},
				{Type: model.ActionCustomCommand, Value: "notify-send morning"},
			},
		},
	}
	loc := mustLocation(t, cfg.Location.Timezone)
	plan, err := ResolveNow(cfg, time.Date(2026, 7, 7, 7, 0, 0, 0, loc))
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.EffectiveActions) != 3 {
		t.Fatalf("effective actions = %#v, want theme, video, and transition command", plan.EffectiveActions)
	}
}

func effectiveActionsByKey(actions []EffectiveAction) map[string]model.Action {
	out := make(map[string]model.Action, len(actions))
	for _, action := range actions {
		out[action.Key] = action.Action
	}
	return out
}

func assertEffectiveAction(t *testing.T, actions map[string]model.Action, key, value string) {
	t.Helper()
	action, ok := actions[key]
	if !ok {
		t.Fatalf("effective action %q is missing from %#v", key, actions)
	}
	if action.Value != value {
		t.Fatalf("effective action %q value = %q, want %q", key, action.Value, value)
	}
}

func mustLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func sunsetForDate(t *testing.T, cfg model.Config, date time.Time) time.Time {
	t.Helper()
	events, err := ResolveSolarEvents(cfg, date)
	if err != nil {
		t.Fatal(err)
	}
	for _, event := range events {
		if event.Kind == model.TriggerSunset {
			if event.Err != nil {
				t.Fatal(event.Err)
			}
			return event.At
		}
	}
	t.Fatal("sunset event not found")
	return time.Time{}
}

func minuteOfDayForTest(at time.Time) int {
	return at.Hour()*60 + at.Minute()
}
