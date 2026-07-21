package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/themetime/themetime/internal/model"
)

const appDir = "themetime"

type Paths struct {
	ConfigDir string
	StateDir  string
	Config    string
	State     string
	Snapshots string
}

func UserPaths() (Paths, error) {
	configBase, err := os.UserConfigDir()
	if err != nil {
		return Paths{}, err
	}
	stateBase := os.Getenv("XDG_STATE_HOME")
	if stateBase == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Paths{}, err
		}
		stateBase = filepath.Join(home, ".local", "state")
	}
	configDir := filepath.Join(configBase, appDir)
	stateDir := filepath.Join(stateBase, appDir)
	return Paths{
		ConfigDir: configDir,
		StateDir:  stateDir,
		Config:    filepath.Join(configDir, "config.json"),
		State:     filepath.Join(stateDir, "state.json"),
		Snapshots: filepath.Join(stateDir, "snapshots"),
	}, nil
}

func RootPaths() Paths {
	return Paths{
		ConfigDir: "/etc/themetime",
		StateDir:  "/var/lib/themetime",
		Config:    "/etc/themetime/privileged-schedule.json",
		State:     "/var/lib/themetime/state.json",
		Snapshots: "/var/lib/themetime/snapshots",
	}
}

func LoadOrCreateDefault() (model.Config, Paths, error) {
	paths, err := UserPaths()
	if err != nil {
		return model.Config{}, Paths{}, err
	}
	cfg, err := Load(paths.Config)
	if err == nil {
		return cfg, paths, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return model.Config{}, paths, err
	}
	cfg = model.DefaultConfig()
	if err := Save(paths.Config, cfg); err != nil {
		return model.Config{}, paths, err
	}
	return cfg, paths, nil
}

func Load(path string) (model.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.Config{}, err
	}
	var cfg model.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return model.Config{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if cfg.Version == 0 {
		cfg.Version = model.CurrentConfigVersion
	}
	if err := cfg.Validate(); err != nil {
		return model.Config{}, err
	}
	return cfg, nil
}

func Save(path string, cfg model.Config) error {
	if cfg.Version == 0 {
		cfg.Version = model.CurrentConfigVersion
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
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

func SnapshotFile(snapshotDir, path string) (string, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(snapshotDir, 0o700); err != nil {
		return "", err
	}
	base := filepath.Base(path)
	name := fmt.Sprintf("%s.%s.bak", time.Now().Format("20060102-150405"), base)
	out := filepath.Join(snapshotDir, name)
	if err := os.WriteFile(out, data, 0o600); err != nil {
		return "", err
	}
	return out, nil
}
