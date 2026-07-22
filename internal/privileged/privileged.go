package privileged

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/RedUndercover/themetime/internal/config"
	"github.com/RedUndercover/themetime/internal/kde"
	"github.com/RedUndercover/themetime/internal/model"
	"github.com/RedUndercover/themetime/internal/scheduler"
)

var safeThemeID = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

type Schedule struct {
	Version int          `json:"version"`
	UserUID string       `json:"userUid"`
	Config  model.Config `json:"config"`
	Written time.Time    `json:"written"`
}

type ApplyResult struct {
	PhaseID string `json:"phaseId"`
	Action  string `json:"action"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type State struct {
	LastPhaseID     string                        `json:"lastPhaseId"`
	LastFingerprint string                        `json:"lastFingerprint"`
	LastAppliedAt   time.Time                     `json:"lastAppliedAt"`
	AppliedActions  map[string]AppliedActionState `json:"appliedActions,omitempty"`
}

type AppliedActionState struct {
	PhaseFingerprint  string    `json:"phaseFingerprint"`
	ActionFingerprint string    `json:"actionFingerprint"`
	AppliedAt         time.Time `json:"appliedAt"`
}

func FilterConfig(cfg model.Config) model.Config {
	out := cfg
	out.Phases = nil
	for _, phase := range cfg.Phases {
		filtered := phase
		filtered.Actions = nil
		for _, action := range phase.Actions {
			if action.IsPrivileged() {
				filtered.Actions = append(filtered.Actions, action)
			}
		}
		if len(filtered.Actions) > 0 {
			out.Phases = append(out.Phases, filtered)
		}
	}
	return out
}

func ValidateSchedule(schedule Schedule) error {
	if schedule.Version <= 0 {
		return errors.New("version is required")
	}
	if strings.TrimSpace(schedule.UserUID) == "" {
		return errors.New("user uid is required")
	}
	if err := schedule.Config.Validate(); err != nil {
		return err
	}
	for _, phase := range schedule.Config.Phases {
		for _, action := range phase.Actions {
			if !action.IsPrivileged() {
				return fmt.Errorf("non-privileged action %q is not allowed in root schedule", action.Type)
			}
			if !safeThemeID.MatchString(action.Value) || strings.Contains(action.Value, "..") {
				return fmt.Errorf("unsafe theme id %q", action.Value)
			}
		}
	}
	return nil
}

func InstallSchedule(path string, schedule Schedule) error {
	if err := ValidateSchedule(schedule); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(schedule, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := fmt.Sprintf("%s.%d.tmp", path, time.Now().UnixNano())
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func LoadSchedule(path string) (Schedule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Schedule{}, err
	}
	var schedule Schedule
	if err := json.Unmarshal(data, &schedule); err != nil {
		return Schedule{}, err
	}
	return schedule, ValidateSchedule(schedule)
}

func ApplyDue(ctx context.Context, r kde.Runner, schedule Schedule, now time.Time) ([]ApplyResult, error) {
	if !schedule.Config.Runtime.Enabled {
		return nil, nil
	}
	plan, err := scheduler.ResolveNow(schedule.Config, now)
	if err != nil {
		return nil, err
	}
	if plan.Active == nil {
		return nil, nil
	}
	var results []ApplyResult
	for _, action := range plan.Active.Phase.Actions {
		result := ApplyResult{PhaseID: plan.Active.Phase.ID, Action: string(action.Type)}
		if err := ApplyAction(ctx, r, action); err != nil {
			result.Error = err.Error()
		} else {
			result.Message = "applied"
		}
		results = append(results, result)
	}
	return results, nil
}

func ApplyDueStateful(ctx context.Context, r kde.Runner, schedule Schedule, statePath string, now time.Time) ([]ApplyResult, error) {
	if !schedule.Config.Runtime.Enabled {
		return nil, nil
	}
	plan, err := scheduler.ResolveNow(schedule.Config, now)
	if err != nil {
		return nil, err
	}
	if plan.Active == nil {
		return nil, nil
	}
	state, _ := LoadState(statePath)
	fingerprint := scheduler.PhaseFingerprint(plan.Active.Phase)
	if state.LastPhaseID == plan.Active.Phase.ID && state.LastFingerprint == fingerprint {
		return nil, nil
	}
	if state.AppliedActions == nil {
		state.AppliedActions = map[string]AppliedActionState{}
	}
	var results []ApplyResult
	failed := false
	for i, action := range plan.Active.Phase.Actions {
		key := appliedActionKey(plan.Active.Phase, i, action)
		actionFingerprint := fingerprintAction(action)
		applied := state.AppliedActions[key]
		if applied.PhaseFingerprint == fingerprint && applied.ActionFingerprint == actionFingerprint {
			continue
		}

		result := ApplyResult{PhaseID: plan.Active.Phase.ID, Action: string(action.Type)}
		if err := ApplyAction(ctx, r, action); err != nil {
			result.Error = err.Error()
			failed = true
		} else {
			result.Message = "applied"
			state.AppliedActions[key] = AppliedActionState{
				PhaseFingerprint:  fingerprint,
				ActionFingerprint: actionFingerprint,
				AppliedAt:         now,
			}
			if err := SaveState(statePath, state); err != nil {
				result.Error = err.Error()
				failed = true
			}
		}
		results = append(results, result)
	}

	if failed {
		return results, fmt.Errorf("one or more privileged actions failed for phase %q", plan.Active.Phase.ID)
	}
	state = State{
		LastPhaseID:     plan.Active.Phase.ID,
		LastFingerprint: fingerprint,
		LastAppliedAt:   now,
	}
	if err := SaveState(statePath, state); err != nil {
		return results, err
	}
	return results, nil
}

func LoadState(path string) (State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return State{}, err
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, err
	}
	return state, nil
}

func SaveState(path string, state State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := fmt.Sprintf("%s.%d.tmp", path, time.Now().UnixNano())
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func ApplyAction(ctx context.Context, r kde.Runner, action model.Action) error {
	switch action.Type {
	case model.ActionSDDMTheme:
		return applySDDMTheme(action.Value)
	case model.ActionPlymouthTheme:
		return applyPlymouthTheme(ctx, r, action.Value)
	default:
		return fmt.Errorf("root helper does not support %q", action.Type)
	}
}

func applySDDMTheme(theme string) error {
	if strings.TrimSpace(theme) == "" {
		return errors.New("sddm theme is required")
	}
	if !safeThemeID.MatchString(theme) || strings.Contains(theme, "..") {
		return fmt.Errorf("unsafe sddm theme id %q", theme)
	}
	if !sddmThemeExists(theme) {
		return fmt.Errorf("sddm theme %q is not installed", theme)
	}
	paths := config.RootPaths()
	if _, err := config.SnapshotFile(paths.Snapshots, "/etc/sddm.conf.d/90-themetime.conf"); err != nil {
		return err
	}
	if err := os.MkdirAll("/etc/sddm.conf.d", 0o755); err != nil {
		return err
	}
	content := fmt.Sprintf("[Theme]\nCurrent=%s\n", theme)
	return os.WriteFile("/etc/sddm.conf.d/90-themetime.conf", []byte(content), 0o644)
}

func applyPlymouthTheme(ctx context.Context, r kde.Runner, theme string) error {
	if strings.TrimSpace(theme) == "" {
		return errors.New("plymouth theme is required")
	}
	if !safeThemeID.MatchString(theme) || strings.Contains(theme, "..") {
		return fmt.Errorf("unsafe plymouth theme id %q", theme)
	}
	if _, err := r.LookPath("plymouth-set-default-theme"); err != nil {
		return errors.New("plymouth-set-default-theme is not installed on this distribution")
	}
	out, err := r.Run(ctx, "plymouth-set-default-theme", "--list")
	if err != nil {
		return err
	}
	if !lineExists(out, theme) {
		return fmt.Errorf("plymouth theme %q is not installed", theme)
	}
	_, err = r.Run(ctx, "plymouth-set-default-theme", "-R", theme)
	return err
}

func sddmThemeExists(theme string) bool {
	for _, root := range []string{"/usr/share/sddm/themes", "/usr/local/share/sddm/themes"} {
		if info, err := os.Stat(filepath.Join(root, theme)); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

func lineExists(out, value string) bool {
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == value {
			return true
		}
	}
	return false
}

func appliedActionKey(phase model.Phase, index int, action model.Action) string {
	return phase.ID + "\n" + strconv.Itoa(index) + "\n" + string(action.Type)
}

func fingerprintAction(action model.Action) string {
	data, _ := json.Marshal(action)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
