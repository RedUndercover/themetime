package daemon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RedUndercover/themetime/internal/config"
	"github.com/RedUndercover/themetime/internal/jsonfile"
	"github.com/RedUndercover/themetime/internal/kde"
	"github.com/RedUndercover/themetime/internal/model"
	"github.com/RedUndercover/themetime/internal/scheduler"
)

type Options struct {
	ConfigPath  string
	StatePath   string
	SnapshotDir string
	PollEvery   time.Duration
	Once        bool
}

type State struct {
	LastPhaseID      string            `json:"lastPhaseId"`
	LastFingerprint  string            `json:"lastFingerprint"`
	LastTransitionAt time.Time         `json:"lastTransitionAt,omitempty"`
	LastAppliedAt    time.Time         `json:"lastAppliedAt"`
	AppliedActions   map[string]string `json:"appliedActions,omitempty"`
}

func Run(ctx context.Context, opts Options) error {
	if opts.PollEvery == 0 {
		opts.PollEvery = 15 * time.Second
	}
	paths, err := config.UserPaths()
	if err != nil {
		return err
	}
	if opts.ConfigPath == "" {
		opts.ConfigPath = paths.Config
	}
	if opts.StatePath == "" {
		opts.StatePath = paths.State
	}
	if opts.SnapshotDir == "" {
		opts.SnapshotDir = paths.Snapshots
	}
	applier := kde.NewApplier(opts.SnapshotDir)
	startup := true
	ticker := time.NewTicker(opts.PollEvery)
	defer ticker.Stop()

	for {
		err := tick(ctx, opts, applier, startup)
		if opts.Once {
			return err
		}
		startup = false
		if err != nil {
			log.Printf("tick failed: %v", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func tick(ctx context.Context, opts Options, applier kde.Applier, startup bool) error {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg = model.DefaultConfig()
			if err := config.Save(opts.ConfigPath, cfg); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	if !cfg.Runtime.Enabled {
		return nil
	}
	plan, err := scheduler.ResolveNow(cfg, time.Now())
	if err != nil {
		return err
	}
	if plan.Active == nil {
		return nil
	}
	effectivePhase := plan.EffectivePhase()
	if effectivePhase == nil {
		return nil
	}
	state, _ := loadState(opts.StatePath)
	fingerprint := scheduler.PhaseFingerprint(*effectivePhase)
	reapplyStartup := startup && cfg.Runtime.ReapplyOnStartup
	if !reapplyStartup && state.LastPhaseID == plan.Active.Phase.ID && state.LastFingerprint == fingerprint && state.LastTransitionAt.Equal(plan.Active.At) {
		return nil
	}
	pendingPhase, appliedActions := pendingEffectiveActions(plan, state, reapplyStartup)
	results := applier.ApplyPhase(ctx, pendingPhase)
	failed := false
	for _, result := range results {
		if result.Error != "" {
			failed = true
			log.Printf("apply %s failed: %s", result.Action.Type, result.Error)
		} else if result.Skipped {
			log.Printf("apply %s skipped: %s", result.Action.Type, result.Message)
		}
	}
	if failed {
		return fmt.Errorf("one or more actions failed for phase %q", plan.Active.Phase.ID)
	}
	state = State{
		LastPhaseID:      plan.Active.Phase.ID,
		LastFingerprint:  fingerprint,
		LastTransitionAt: plan.Active.At,
		LastAppliedAt:    time.Now(),
		AppliedActions:   appliedActions,
	}
	return saveState(opts.StatePath, state)
}

func pendingEffectiveActions(plan scheduler.Plan, state State, reapply bool) (model.Phase, map[string]string) {
	phase := *plan.EffectivePhase()
	phase.Actions = nil
	desired := map[string]string{}
	transitionChanged := state.LastTransitionAt.IsZero() || !state.LastTransitionAt.Equal(plan.Active.At)

	for _, effective := range plan.EffectiveActions {
		if effective.Event {
			if transitionChanged {
				phase.Actions = append(phase.Actions, effective.Action)
			}
			continue
		}
		fingerprint := scheduler.ActionFingerprint(effective.Action)
		desired[effective.Key] = fingerprint
		if reapply || state.AppliedActions == nil || state.AppliedActions[effective.Key] != fingerprint {
			phase.Actions = append(phase.Actions, effective.Action)
		}
	}
	return phase, desired
}

func loadState(path string) (State, error) {
	var state State
	if err := jsonfile.Read(path, &state); err != nil {
		return State{}, err
	}
	return state, nil
}

func saveState(path string, state State) error {
	return jsonfile.WriteAtomic(path, state)
}

func ApplyPhaseByID(ctx context.Context, cfg model.Config, snapshotDir string, phaseID string) ([]kde.Result, error) {
	for _, phase := range cfg.Phases {
		if phase.ID == phaseID {
			applier := kde.NewApplier(snapshotDir)
			return applier.ApplyPhase(ctx, phase), nil
		}
	}
	return nil, fmt.Errorf("phase %q not found", phaseID)
}

func ApplyEffectiveNow(ctx context.Context, cfg model.Config, snapshotDir string, now time.Time) (model.Phase, []kde.Result, error) {
	plan, err := scheduler.ResolveNow(cfg, now)
	if err != nil {
		return model.Phase{}, nil, err
	}
	phase := plan.EffectivePhase()
	if phase == nil {
		return model.Phase{}, nil, errors.New("no active phase is available")
	}
	applier := kde.NewApplier(snapshotDir)
	return *phase, applier.ApplyPhase(ctx, *phase), nil
}
