package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/themetime/themetime/internal/config"
	"github.com/themetime/themetime/internal/daemon"
	"github.com/themetime/themetime/internal/doctor"
	"github.com/themetime/themetime/internal/kde"
	"github.com/themetime/themetime/internal/model"
	"github.com/themetime/themetime/internal/scheduler"
	"github.com/themetime/themetime/internal/systemd"
)

//go:embed all:frontend/dist
var assets embed.FS

var hexColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

type App struct {
	mu    sync.Mutex
	ctx   context.Context
	cfg   model.Config
	paths config.Paths
	inv   kde.Inventory
}

type UIState struct {
	Config        model.Config       `json:"config"`
	Inventory     kde.Inventory      `json:"inventory"`
	Paths         config.Paths       `json:"paths"`
	Now           time.Time          `json:"now"`
	Today         string             `json:"today"`
	Plan          PlanView           `json:"plan"`
	SolarEvents   []SolarEventView   `json:"solarEvents"`
	Transitions   []TransitionView   `json:"transitions"`
	ActionOptions []ActionOptionView `json:"actionOptions"`
	Checks        []doctor.Check     `json:"checks"`
}

type PlanView struct {
	Active *TransitionView `json:"active,omitempty"`
	Next   *TransitionView `json:"next,omitempty"`
}

type TransitionView struct {
	PhaseID   string    `json:"phaseId"`
	PhaseName string    `json:"phaseName"`
	Clock     string    `json:"clock"`
	At        time.Time `json:"at"`
	Color     string    `json:"color"`
	Trigger   string    `json:"trigger"`
}

type SolarEventView struct {
	Kind       model.TriggerKind `json:"kind"`
	Label      string            `json:"label"`
	ShortLabel string            `json:"shortLabel"`
	At         time.Time         `json:"at,omitempty"`
	Clock      string            `json:"clock,omitempty"`
	Error      string            `json:"error,omitempty"`
}

type ActionOptionView struct {
	Type        model.ActionType `json:"type"`
	Label       string           `json:"label"`
	Placeholder string           `json:"placeholder"`
	Choices     []kde.Choice     `json:"choices"`
	Warning     string           `json:"warning,omitempty"`
}

type LocationPreviewView struct {
	Today       string           `json:"today"`
	SolarEvents []SolarEventView `json:"solarEvents"`
	Transitions []TransitionView `json:"transitions"`
}

const applicationID = "io.github.themetime.ThemeTime"

func main() {
	app := &App{}
	tray := newTrayController(app)
	startTray, stopTray := tray.externalLoop()
	startTray()
	defer stopTray()
	err := wails.Run(&options.App{
		Title:     "ThemeTime",
		Width:     1440,
		Height:    900,
		MinWidth:  980,
		MinHeight: 680,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour:  options.NewRGBA(5, 14, 25, 255),
		HideWindowOnClose: true,
		OnStartup:         app.startup,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: applicationID,
			OnSecondInstanceLaunch: func(options.SecondInstanceData) {
				app.showWindow()
			},
		},
		Linux: &linux.Options{
			Icon:        themeTimeIconPNG(),
			ProgramName: applicationID,
		},
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (a *App) startup(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ctx = ctx
}

func (a *App) GetState() (UIState, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.reload(); err != nil {
		return UIState{}, err
	}
	return a.state()
}

func (a *App) SaveConfig(next model.Config) (UIState, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.paths.Config == "" {
		if err := a.reload(); err != nil {
			return UIState{}, err
		}
	}
	next = normalizeMediaPaths(next)
	if err := a.validateConfig(next); err != nil {
		return UIState{}, err
	}
	if err := config.Save(a.paths.Config, next); err != nil {
		return UIState{}, err
	}
	a.cfg = next
	return a.state()
}

func normalizeMediaPaths(cfg model.Config) model.Config {
	out := cfg
	out.Phases = append([]model.Phase(nil), cfg.Phases...)
	for phaseIndex := range out.Phases {
		out.Phases[phaseIndex].Actions = append([]model.Action(nil), cfg.Phases[phaseIndex].Actions...)
		for actionIndex := range out.Phases[phaseIndex].Actions {
			action := &out.Phases[phaseIndex].Actions[actionIndex]
			if action.Type != model.ActionStaticWallpaper && action.Type != model.ActionVideoWallpaper {
				continue
			}
			local := localFilePath(action.Value)
			if local == "" {
				continue
			}
			absolute, err := filepath.Abs(local)
			if err == nil {
				action.Value = absolute
			}
		}
	}
	return out
}

func (a *App) ApplyPhase(id string) ([]kde.Result, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.paths.Snapshots == "" {
		if err := a.reload(); err != nil {
			return nil, err
		}
	}
	return daemon.ApplyPhaseByID(context.Background(), a.cfg, a.paths.Snapshots, id)
}

func (a *App) PreviewLocation(location model.Location) (LocationPreviewView, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.paths.Config == "" {
		if err := a.reload(); err != nil {
			return LocationPreviewView{}, err
		}
	}
	return previewLocation(a.cfg, location, time.Now())
}

func (a *App) RunDoctor() []doctor.Check {
	return doctor.Run(context.Background())
}

func (a *App) InstallUserService() (string, error) {
	return systemd.InstallUserService("", true)
}

func (a *App) showWindow() {
	a.mu.Lock()
	ctx := a.ctx
	a.mu.Unlock()
	if ctx == nil {
		return
	}
	wailsruntime.WindowShow(ctx)
	wailsruntime.WindowUnminimise(ctx)
}

func (a *App) quit() {
	a.mu.Lock()
	ctx := a.ctx
	a.mu.Unlock()
	if ctx != nil {
		wailsruntime.Quit(ctx)
	}
}

func (a *App) scheduleSummary() (active string, next string, err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.paths.Config == "" {
		if err := a.reload(); err != nil {
			return "", "", err
		}
	}
	plan, err := scheduler.ResolveNow(a.cfg, time.Now())
	if err != nil {
		return "", "", err
	}
	if plan.Active != nil {
		active = plan.Active.Phase.Name
	}
	if plan.Next != nil {
		next = plan.Next.Phase.Name + " " + transitionClockLabel(a.cfg, plan.Next.At, time.Now())
	}
	return active, next, nil
}

func (a *App) applyCurrent() (string, []kde.Result, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.paths.Snapshots == "" {
		if err := a.reload(); err != nil {
			return "", nil, err
		}
	}
	phase, results, err := daemon.ApplyEffectiveNow(context.Background(), a.cfg, a.paths.Snapshots, time.Now())
	return phase.Name, results, err
}

func (a *App) reload() error {
	cfg, paths, err := config.LoadOrCreateDefault()
	if err != nil {
		return err
	}
	a.cfg = cfg
	a.paths = paths
	a.inv = kde.Discover(context.Background(), kde.ExecRunner{})
	return nil
}

func (a *App) state() (UIState, error) {
	now := time.Now()
	plan, err := scheduler.ResolveNow(a.cfg, now)
	if err != nil {
		return UIState{}, err
	}
	events, err := scheduler.ResolveSolarEvents(a.cfg, now)
	if err != nil {
		return UIState{}, err
	}
	transitions, err := scheduler.ResolveDay(a.cfg, now)
	if err != nil {
		return UIState{}, err
	}
	return UIState{
		Config:        a.cfg,
		Inventory:     a.inv,
		Paths:         a.paths,
		Now:           now,
		Today:         scheduleDayLabel(a.cfg, now),
		Plan:          planView(a.cfg, plan, now),
		SolarEvents:   solarEventViews(events),
		Transitions:   transitionViews(a.cfg, transitions, now),
		ActionOptions: a.actionOptions(),
		Checks:        doctor.Run(context.Background()),
	}, nil
}

func previewLocation(cfg model.Config, location model.Location, now time.Time) (LocationPreviewView, error) {
	location.Source = "manual"
	cfg.Location = location
	if err := cfg.Validate(); err != nil {
		return LocationPreviewView{}, err
	}
	events, err := scheduler.ResolveSolarEvents(cfg, now)
	if err != nil {
		return LocationPreviewView{}, err
	}
	transitions, err := scheduler.ResolveDay(cfg, now)
	if err != nil {
		return LocationPreviewView{}, err
	}
	return LocationPreviewView{
		Today:       scheduleDayLabel(cfg, now),
		SolarEvents: solarEventViews(events),
		Transitions: transitionViews(cfg, transitions, now),
	}, nil
}

func (a *App) validateConfig(cfg model.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	for _, phase := range cfg.Phases {
		if strings.TrimSpace(phase.Color) != "" && !hexColorPattern.MatchString(strings.TrimSpace(phase.Color)) {
			return fmt.Errorf("phase %q color %q must be #RRGGBB", phase.Name, phase.Color)
		}
		for _, action := range phase.Actions {
			if err := a.validateAction(action); err != nil {
				return fmt.Errorf("phase %q: %w", phase.Name, err)
			}
		}
	}
	return nil
}

func (a *App) validateAction(action model.Action) error {
	if err := action.Validate(); err != nil {
		return err
	}
	switch action.Type {
	case model.ActionColorScheme:
		return validateChoice(action, a.inv.ColorSchemes, "color scheme")
	case model.ActionGlobalTheme:
		return validateChoice(action, a.inv.GlobalThemes, "global theme")
	case model.ActionAccentColor:
		if !hexColorPattern.MatchString(action.Value) {
			return fmt.Errorf("accent color %q must be #RRGGBB", action.Value)
		}
	case model.ActionPlasmaStyle:
		return validateChoice(action, a.inv.PlasmaStyles, "Plasma style")
	case model.ActionIconTheme:
		return validateChoice(action, a.inv.IconThemes, "icon theme")
	case model.ActionCursorTheme:
		return validateChoice(action, a.inv.CursorThemes, "cursor theme")
	case model.ActionWindowDecoration:
		return validateChoice(action, a.inv.WindowDecorations, "window decoration")
	case model.ActionStaticWallpaper:
		return validateExistingFile(action.Value, "static wallpaper", staticWallpaperExtensions())
	case model.ActionVideoWallpaper:
		if !a.inv.SmartVideoPlugin {
			return errors.New("video wallpaper requires Smart Video Wallpaper Reborn")
		}
		return validateExistingFile(action.Value, "video wallpaper", videoWallpaperExtensions())
	case model.ActionSDDMTheme:
		return validateChoice(action, a.inv.SDDMThemes, "SDDM theme")
	case model.ActionPlymouthTheme:
		return validateChoice(action, a.inv.PlymouthThemes, "Plymouth theme")
	}
	return nil
}

func (a *App) actionOptions() []ActionOptionView {
	add := func(action model.ActionType, choices []kde.Choice, placeholder string, warning string) ActionOptionView {
		return ActionOptionView{Type: action, Label: model.Action{Type: action}.Label(), Placeholder: placeholder, Choices: choices, Warning: warning}
	}
	options := []ActionOptionView{
		add(model.ActionAccentColor, nil, "#3daee9", ""),
		add(model.ActionStaticWallpaper, a.inv.StaticWallpapers, "/path/to/image.jpg", ""),
	}
	videoWarning := ""
	if !a.inv.SmartVideoPlugin {
		videoWarning = "Smart Video Wallpaper Reborn is not installed yet."
	}
	options = append(options, add(model.ActionVideoWallpaper, a.inv.VideoWallpapers, "/path/to/video.mp4", videoWarning))
	appendIf := func(ok bool, action model.ActionType, choices []kde.Choice) {
		if ok {
			options = append(options, add(action, choices, "installed value", ""))
		}
	}
	appendIf(len(a.inv.ColorSchemes) > 0, model.ActionColorScheme, a.inv.ColorSchemes)
	appendIf(len(a.inv.GlobalThemes) > 0, model.ActionGlobalTheme, a.inv.GlobalThemes)
	appendIf(len(a.inv.PlasmaStyles) > 0, model.ActionPlasmaStyle, a.inv.PlasmaStyles)
	appendIf(len(a.inv.IconThemes) > 0, model.ActionIconTheme, a.inv.IconThemes)
	appendIf(len(a.inv.CursorThemes) > 0, model.ActionCursorTheme, a.inv.CursorThemes)
	appendIf(len(a.inv.WindowDecorations) > 0, model.ActionWindowDecoration, a.inv.WindowDecorations)
	appendIf(len(a.inv.SDDMThemes) > 0, model.ActionSDDMTheme, a.inv.SDDMThemes)
	appendIf(len(a.inv.PlymouthThemes) > 0, model.ActionPlymouthTheme, a.inv.PlymouthThemes)
	return options
}

func planView(cfg model.Config, plan scheduler.Plan, now time.Time) PlanView {
	var out PlanView
	if plan.Active != nil {
		active := transitionView(cfg, *plan.Active, now)
		out.Active = &active
	}
	if plan.Next != nil {
		next := transitionView(cfg, *plan.Next, now)
		out.Next = &next
	}
	return out
}

func transitionViews(cfg model.Config, transitions []scheduler.Transition, now time.Time) []TransitionView {
	out := make([]TransitionView, 0, len(transitions))
	for _, transition := range transitions {
		out = append(out, transitionView(cfg, transition, now))
	}
	return out
}

func transitionView(cfg model.Config, transition scheduler.Transition, now time.Time) TransitionView {
	return TransitionView{
		PhaseID:   transition.Phase.ID,
		PhaseName: transition.Phase.Name,
		Clock:     transitionClockLabel(cfg, transition.At, now),
		At:        transition.At,
		Color:     transition.Phase.Color,
		Trigger:   transition.Phase.Start.String(),
	}
}

func solarEventViews(events []scheduler.SolarEvent) []SolarEventView {
	out := make([]SolarEventView, 0, len(events))
	for _, event := range events {
		view := SolarEventView{
			Kind:       event.Kind,
			Label:      model.SolarTriggerLabel(event.Kind),
			ShortLabel: model.SolarTriggerShortLabel(event.Kind),
			At:         event.At,
		}
		if event.Err != nil {
			view.Error = event.Err.Error()
		}
		if !event.At.IsZero() {
			view.Clock = event.At.Format("15:04")
		}
		out = append(out, view)
	}
	return out
}

func validateChoice(action model.Action, choices []kde.Choice, label string) error {
	if len(choices) == 0 {
		return nil
	}
	value := strings.TrimSpace(action.Value)
	for _, choice := range choices {
		if value == choice.ID {
			return nil
		}
	}
	return fmt.Errorf("%s %q is not installed", label, value)
}

func validateExistingFile(value, label string, extensions []string) error {
	path := localFilePath(value)
	if path == "" {
		return fmt.Errorf("%s path is required", label)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s file %q is not available", label, value)
	}
	if info.IsDir() {
		return fmt.Errorf("%s path %q is a directory", label, value)
	}
	ext := strings.ToLower(filepath.Ext(path))
	if len(extensions) > 0 && !slices.Contains(extensions, ext) {
		return fmt.Errorf("%s file %q has an unsupported extension", label, value)
	}
	return nil
}

func localFilePath(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "file://") {
		uri, err := url.Parse(value)
		if err != nil {
			return ""
		}
		return uri.Path
	}
	if strings.Contains(value, "://") {
		return ""
	}
	if strings.HasPrefix(value, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return value
		}
		if value == "~" {
			return home
		}
		if strings.HasPrefix(value, "~/") {
			return filepath.Join(home, strings.TrimPrefix(value, "~/"))
		}
	}
	return value
}

func staticWallpaperExtensions() []string {
	return []string{".avif", ".bmp", ".jpeg", ".jpg", ".png", ".svg", ".webp"}
}

func videoWallpaperExtensions() []string {
	return []string{".avi", ".m4v", ".mkv", ".mov", ".mp4", ".webm"}
}

func scheduleDayLabel(cfg model.Config, now time.Time) string {
	loc, err := time.LoadLocation(cfg.Location.Timezone)
	if err != nil {
		return "Today"
	}
	return now.In(loc).Format("Mon, Jan 2")
}

func transitionClockLabel(cfg model.Config, at time.Time, now time.Time) string {
	loc, err := time.LoadLocation(cfg.Location.Timezone)
	if err != nil {
		return at.Format("15:04")
	}
	at = at.In(loc)
	now = now.In(loc)
	if sameLocalDay(at, now) {
		return at.Format("15:04")
	}
	if sameLocalDay(at, now.AddDate(0, 0, 1)) {
		return "Tomorrow " + at.Format("15:04")
	}
	return at.Format("Jan 2 15:04")
}

func sameLocalDay(a time.Time, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
