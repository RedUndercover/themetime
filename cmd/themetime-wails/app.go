package main

import (
	"context"
	"sync"
	"time"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/RedUndercover/themetime/internal/config"
	"github.com/RedUndercover/themetime/internal/daemon"
	"github.com/RedUndercover/themetime/internal/doctor"
	"github.com/RedUndercover/themetime/internal/kde"
	"github.com/RedUndercover/themetime/internal/model"
	"github.com/RedUndercover/themetime/internal/scheduler"
	"github.com/RedUndercover/themetime/internal/systemd"
)

type App struct {
	mu    sync.Mutex
	ctx   context.Context
	cfg   model.Config
	paths config.Paths
	inv   kde.Inventory
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

func (a *App) ApplyPhase(id string) ([]kde.Result, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.paths.Snapshots == "" {
		if err := a.reload(); err != nil {
			return nil, err
		}
	}
	return daemon.ApplyPhaseByID(a.contextLocked(), a.cfg, a.paths.Snapshots, id)
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
	return doctor.Run(a.context())
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
	phase, results, err := daemon.ApplyEffectiveNow(a.contextLocked(), a.cfg, a.paths.Snapshots, time.Now())
	return phase.Name, results, err
}

func (a *App) reload() error {
	cfg, paths, err := config.LoadOrCreateDefault()
	if err != nil {
		return err
	}
	a.cfg = cfg
	a.paths = paths
	a.inv = kde.Discover(a.contextLocked(), kde.ExecRunner{})
	return nil
}

func (a *App) context() context.Context {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.contextLocked()
}

func (a *App) contextLocked() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}
