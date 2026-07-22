package main

import (
	"time"

	"github.com/RedUndercover/themetime/internal/config"
	"github.com/RedUndercover/themetime/internal/doctor"
	"github.com/RedUndercover/themetime/internal/kde"
	"github.com/RedUndercover/themetime/internal/model"
	"github.com/RedUndercover/themetime/internal/scheduler"
)

type UIState struct {
	Config         model.Config              `json:"config"`
	Inventory      kde.Inventory             `json:"inventory"`
	Paths          config.Paths              `json:"paths"`
	Now            time.Time                 `json:"now"`
	Today          string                    `json:"today"`
	Plan           PlanView                  `json:"plan"`
	SolarEvents    []SolarEventView          `json:"solarEvents"`
	Transitions    []TransitionView          `json:"transitions"`
	TriggerOptions []model.TriggerDefinition `json:"triggerOptions"`
	ActionOptions  []ActionOptionView        `json:"actionOptions"`
	Checks         []doctor.Check            `json:"checks"`
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
		Config:         a.cfg,
		Inventory:      a.inv,
		Paths:          a.paths,
		Now:            now,
		Today:          scheduleDayLabel(a.cfg, now),
		Plan:           planView(a.cfg, plan, now),
		SolarEvents:    solarEventViews(events),
		Transitions:    transitionViews(a.cfg, transitions, now),
		TriggerOptions: model.TriggerDefinitions(),
		ActionOptions:  a.actionOptions(),
		Checks:         doctor.Run(a.contextLocked()),
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
