package daemon

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/RedUndercover/themetime/internal/config"
	"github.com/RedUndercover/themetime/internal/kde"
	"github.com/RedUndercover/themetime/internal/model"
	"github.com/RedUndercover/themetime/internal/scheduler"
)

type fakeDaemonRunner struct {
	runs int
}

func (r *fakeDaemonRunner) LookPath(name string) (string, error) {
	return name, nil
}

func (r *fakeDaemonRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	r.runs++
	return "", nil
}

func TestStateJSONRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/state.json"
	want := State{
		LastPhaseID:      "night",
		LastFingerprint:  "abc",
		LastTransitionAt: time.Unix(5, 0),
		LastAppliedAt:    time.Unix(10, 0),
		AppliedActions:   map[string]string{"wallpaper:all": "def"},
	}
	if err := saveState(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := loadState(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.LastPhaseID != want.LastPhaseID || got.LastFingerprint != want.LastFingerprint || !got.LastTransitionAt.Equal(want.LastTransitionAt) || !got.LastAppliedAt.Equal(want.LastAppliedAt) || got.AppliedActions["wallpaper:all"] != "def" {
		t.Fatalf("state = %#v, want %#v", got, want)
	}
}

func TestRunOnceDoesNotSaveStateOnActionFailure(t *testing.T) {
	dir := t.TempDir()
	cfg := model.DefaultConfig()
	cfg.Phases = []model.Phase{
		{
			ID:      "bad",
			Name:    "Bad",
			Color:   "#000000",
			Enabled: true,
			Start:   model.Trigger{Kind: model.TriggerClock, Clock: "00:00"},
			Actions: []model.Action{
				{Type: model.ActionCustomCommand, Value: "exit 2"},
			},
		},
	}
	configPath := dir + "/config.json"
	statePath := dir + "/state.json"
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatal(err)
	}
	if err := Run(context.Background(), Options{ConfigPath: configPath, StatePath: statePath, SnapshotDir: dir, Once: true}); err == nil {
		t.Fatal("expected apply failure")
	}
	if _, err := loadState(statePath); err == nil {
		t.Fatal("state was saved despite action failure")
	}
}

func TestTickReappliesMatchingStateOnStartupWhenConfigured(t *testing.T) {
	dir := t.TempDir()
	cfg := configForStartupReapply(true)
	configPath := dir + "/config.json"
	statePath := dir + "/state.json"
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatal(err)
	}
	if err := saveState(statePath, matchingState(t, cfg)); err != nil {
		t.Fatal(err)
	}

	runner := &fakeDaemonRunner{}
	err := tick(context.Background(), Options{ConfigPath: configPath, StatePath: statePath, SnapshotDir: dir}, kde.Applier{Runner: runner, SnapshotDir: dir}, true)
	if err != nil {
		t.Fatal(err)
	}
	if runner.runs != 1 {
		t.Fatalf("runs = %d, want 1", runner.runs)
	}
}

func TestTickSkipsMatchingStateOnStartupWhenReapplyDisabled(t *testing.T) {
	dir := t.TempDir()
	cfg := configForStartupReapply(false)
	configPath := dir + "/config.json"
	statePath := dir + "/state.json"
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatal(err)
	}
	if err := saveState(statePath, matchingState(t, cfg)); err != nil {
		t.Fatal(err)
	}

	runner := &fakeDaemonRunner{}
	err := tick(context.Background(), Options{ConfigPath: configPath, StatePath: statePath, SnapshotDir: dir}, kde.Applier{Runner: runner, SnapshotDir: dir}, true)
	if err != nil {
		t.Fatal(err)
	}
	if runner.runs != 0 {
		t.Fatalf("runs = %d, want 0", runner.runs)
	}
}

func TestPendingEffectiveActionsOnlyAppliesChangedTrack(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Phases = []model.Phase{
		{
			ID: "theme", Name: "Theme", Enabled: true,
			Start:   model.Trigger{Kind: model.TriggerClock, Clock: "06:00"},
			Actions: []model.Action{{Type: model.ActionColorScheme, Value: "BreezeLight"}},
		},
		{
			ID: "video", Name: "Video", Enabled: true,
			Start:   model.Trigger{Kind: model.TriggerClock, Clock: "07:00"},
			Actions: []model.Action{{Type: model.ActionVideoWallpaper, Value: "/videos/day.mp4"}},
		},
	}
	loc := mustDaemonLocation(t, cfg.Location.Timezone)
	plan, err := scheduler.ResolveNow(cfg, time.Date(2026, 7, 7, 8, 0, 0, 0, loc))
	if err != nil {
		t.Fatal(err)
	}
	color := cfg.Phases[0].Actions[0]
	state := State{
		LastPhaseID:      "theme",
		LastTransitionAt: time.Date(2026, 7, 7, 6, 0, 0, 0, loc),
		AppliedActions:   map[string]string{"action:colorScheme": scheduler.ActionFingerprint(color)},
	}
	pending, desired := pendingEffectiveActions(plan, state, false)
	if len(pending.Actions) != 1 || pending.Actions[0].Type != model.ActionVideoWallpaper {
		t.Fatalf("pending actions = %#v, want only video wallpaper", pending.Actions)
	}
	if len(desired) != 2 {
		t.Fatalf("desired tracks = %#v, want color and wallpaper", desired)
	}
}

func TestPendingEffectiveActionsDoesNotReplayEventOnStartup(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Phases = []model.Phase{
		{
			ID: "command", Name: "Command", Enabled: true,
			Start:   model.Trigger{Kind: model.TriggerClock, Clock: "06:00"},
			Actions: []model.Action{{Type: model.ActionCustomCommand, Value: "notify-send morning"}},
		},
	}
	loc := mustDaemonLocation(t, cfg.Location.Timezone)
	plan, err := scheduler.ResolveNow(cfg, time.Date(2026, 7, 7, 8, 0, 0, 0, loc))
	if err != nil {
		t.Fatal(err)
	}
	state := State{LastPhaseID: "command", LastTransitionAt: plan.Active.At}
	pending, _ := pendingEffectiveActions(plan, state, true)
	if len(pending.Actions) != 0 {
		t.Fatalf("startup replayed event actions: %#v", pending.Actions)
	}
}

func matchingState(t *testing.T, cfg model.Config) State {
	t.Helper()
	plan, err := scheduler.ResolveNow(cfg, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	phase := plan.EffectivePhase()
	applied := map[string]string{}
	for _, action := range plan.EffectiveActions {
		if !action.Event {
			applied[action.Key] = scheduler.ActionFingerprint(action.Action)
		}
	}
	return State{
		LastPhaseID:      plan.Active.Phase.ID,
		LastFingerprint:  scheduler.PhaseFingerprint(*phase),
		LastTransitionAt: plan.Active.At,
		LastAppliedAt:    time.Now(),
		AppliedActions:   applied,
	}
}

func mustDaemonLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func configForStartupReapply(reapply bool) model.Config {
	cfg := model.DefaultConfig()
	cfg.Runtime.ReapplyOnStartup = reapply
	cfg.Phases = []model.Phase{
		{
			ID:      "day",
			Name:    "Day",
			Color:   "#ffffff",
			Enabled: true,
			Start:   model.Trigger{Kind: model.TriggerClock, Clock: "00:00"},
			Actions: []model.Action{
				{Type: model.ActionColorScheme, Value: fmt.Sprintf("BreezeLight-%t", reapply)},
			},
		},
	}
	return cfg
}
