package privileged

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/themetime/themetime/internal/model"
)

type fakePrivilegedRunner struct {
	rebuilds int
	applied  []string
	failures map[string]bool
}

func (r *fakePrivilegedRunner) LookPath(name string) (string, error) {
	if name == "plymouth-set-default-theme" {
		return "/usr/bin/plymouth-set-default-theme", nil
	}
	return "", fmt.Errorf("%s not found", name)
}

func (r *fakePrivilegedRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	if name == "plymouth-set-default-theme" && len(args) == 1 && args[0] == "--list" {
		return "spinner\nbgrt\n", nil
	}
	if name == "plymouth-set-default-theme" && len(args) == 2 && args[0] == "-R" {
		theme := args[1]
		if r.failures[theme] {
			return "", fmt.Errorf("failed to apply %s", theme)
		}
		r.applied = append(r.applied, theme)
		if theme == "spinner" {
			r.rebuilds++
		}
		return "", nil
	}
	return "", fmt.Errorf("unexpected command: %s %v", name, args)
}

func TestValidateScheduleRejectsNonPrivilegedAction(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Phases = []model.Phase{
		{
			ID:      "bad",
			Name:    "Bad",
			Color:   "#000000",
			Enabled: true,
			Start:   model.Trigger{Kind: model.TriggerClock, Clock: "12:00"},
			Actions: []model.Action{
				{Type: model.ActionCustomCommand, Value: "touch /tmp/nope"},
			},
		},
	}
	err := ValidateSchedule(Schedule{Version: 1, UserUID: "1000", Config: cfg, Written: time.Now()})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateScheduleRejectsUnsafeThemeID(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Phases = []model.Phase{
		{
			ID:      "bad",
			Name:    "Bad",
			Color:   "#000000",
			Enabled: true,
			Start:   model.Trigger{Kind: model.TriggerClock, Clock: "12:00"},
			Actions: []model.Action{
				{Type: model.ActionSDDMTheme, Value: "breeze\n[Users]\nMaximumUid=0"},
			},
		},
	}
	err := ValidateSchedule(Schedule{Version: 1, UserUID: "1000", Config: cfg, Written: time.Now()})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestFilterConfigKeepsOnlyPrivilegedActions(t *testing.T) {
	cfg := model.DefaultConfig()
	cfg.Phases[0].Actions = append(cfg.Phases[0].Actions, model.Action{Type: model.ActionSDDMTheme, Value: "breeze"})
	filtered := FilterConfig(cfg)
	if len(filtered.Phases) != 1 {
		t.Fatalf("phases = %d, want 1", len(filtered.Phases))
	}
	if len(filtered.Phases[0].Actions) != 1 || filtered.Phases[0].Actions[0].Type != model.ActionSDDMTheme {
		t.Fatalf("actions = %#v", filtered.Phases[0].Actions)
	}
}

func TestApplyDueStatefulSkipsAlreadyAppliedFingerprint(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	cfg := model.DefaultConfig()
	cfg.Phases = []model.Phase{
		{
			ID:      "boot",
			Name:    "Boot",
			Color:   "#000000",
			Enabled: true,
			Start:   model.Trigger{Kind: model.TriggerClock, Clock: "00:00"},
			Actions: []model.Action{
				{Type: model.ActionPlymouthTheme, Value: "spinner"},
			},
		},
	}
	schedule := Schedule{Version: 1, UserUID: "1000", Config: cfg, Written: now}
	runner := &fakePrivilegedRunner{}
	statePath := t.TempDir() + "/state.json"

	results, err := ApplyDueStateful(context.Background(), runner, schedule, statePath, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	if runner.rebuilds != 1 {
		t.Fatalf("rebuilds = %d, want 1", runner.rebuilds)
	}
	state, err := LoadState(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if state.LastPhaseID != "boot" || !state.LastAppliedAt.Equal(now) {
		t.Fatalf("state = %#v", state)
	}

	results, err = ApplyDueStateful(context.Background(), runner, schedule, statePath, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("results = %d, want 0", len(results))
	}
	if runner.rebuilds != 1 {
		t.Fatalf("rebuilds = %d, want 1", runner.rebuilds)
	}
}

func TestApplyDueStatefulRetriesOnlyFailedPrivilegedActions(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	cfg := model.DefaultConfig()
	cfg.Phases = []model.Phase{
		{
			ID:      "boot",
			Name:    "Boot",
			Color:   "#000000",
			Enabled: true,
			Start:   model.Trigger{Kind: model.TriggerClock, Clock: "00:00"},
			Actions: []model.Action{
				{Type: model.ActionPlymouthTheme, Value: "spinner"},
				{Type: model.ActionPlymouthTheme, Value: "bgrt"},
			},
		},
	}
	schedule := Schedule{Version: 1, UserUID: "1000", Config: cfg, Written: now}
	runner := &fakePrivilegedRunner{failures: map[string]bool{"bgrt": true}}
	statePath := t.TempDir() + "/state.json"

	results, err := ApplyDueStateful(context.Background(), runner, schedule, statePath, now)
	if err == nil {
		t.Fatal("expected bgrt failure")
	}
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
	if runner.rebuilds != 1 {
		t.Fatalf("rebuilds = %d, want 1", runner.rebuilds)
	}
	state, err := LoadState(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if state.LastPhaseID != "" {
		t.Fatalf("phase was marked complete after partial failure: %#v", state)
	}
	if len(state.AppliedActions) != 1 {
		t.Fatalf("applied action states = %d, want 1", len(state.AppliedActions))
	}

	runner.failures["bgrt"] = false
	results, err = ApplyDueStateful(context.Background(), runner, schedule, statePath, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Action != string(model.ActionPlymouthTheme) {
		t.Fatalf("results = %#v, want only the retried bgrt action", results)
	}
	if runner.rebuilds != 1 {
		t.Fatalf("spinner was rebuilt again; rebuilds = %d, want 1", runner.rebuilds)
	}
	state, err = LoadState(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if state.LastPhaseID != "boot" {
		t.Fatalf("phase was not marked complete: %#v", state)
	}
}

func TestApplyDueStatefulSkipsDisabledRuntime(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	cfg := model.DefaultConfig()
	cfg.Runtime.Enabled = false
	cfg.Phases = []model.Phase{
		{
			ID:      "boot",
			Name:    "Boot",
			Color:   "#000000",
			Enabled: true,
			Start:   model.Trigger{Kind: model.TriggerClock, Clock: "00:00"},
			Actions: []model.Action{
				{Type: model.ActionPlymouthTheme, Value: "spinner"},
			},
		},
	}
	runner := &fakePrivilegedRunner{}
	results, err := ApplyDueStateful(context.Background(), runner, Schedule{Version: 1, UserUID: "1000", Config: cfg, Written: now}, t.TempDir()+"/state.json", now)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("results = %d, want 0", len(results))
	}
	if runner.rebuilds != 0 {
		t.Fatalf("rebuilds = %d, want 0", runner.rebuilds)
	}
}

func TestLineExistsMatchesWholeTrimmedLine(t *testing.T) {
	if !lineExists("spinner\nbgrt\n", "spinner") {
		t.Fatal("expected spinner to exist")
	}
	if lineExists("spinner-dark\nbgrt\n", "spinner") {
		t.Fatal("matched a partial theme name")
	}
}
