package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/themetime/themetime/internal/buildinfo"
	"github.com/themetime/themetime/internal/config"
	"github.com/themetime/themetime/internal/daemon"
	"github.com/themetime/themetime/internal/doctor"
	"github.com/themetime/themetime/internal/privileged"
	"github.com/themetime/themetime/internal/systemd"
)

const rootctlPolicyPath = "/usr/local/libexec/themetime-rootctl"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return runGUI()
	}
	switch os.Args[1] {
	case "gui":
		return runGUI()
	case "daemon":
		return runDaemon(os.Args[2:])
	case "apply":
		return runApply(os.Args[2:])
	case "doctor":
		return runDoctor()
	case "install-user-service":
		return runInstallUserService(os.Args[2:])
	case "install-privileged-schedule":
		return runInstallPrivilegedSchedule()
	case "show-config":
		return runShowConfig()
	case "version", "--version":
		fmt.Println(buildinfo.String())
		return nil
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q", os.Args[1])
	}
}

func runGUI() error {
	cmd, err := guiCommand()
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func runDaemon(args []string) error {
	flags := flag.NewFlagSet("daemon", flag.ExitOnError)
	once := flags.Bool("once", false, "apply once and exit")
	poll := flags.Duration("poll", 15*time.Second, "poll interval")
	configPath := flags.String("config", "", "config path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return daemon.Run(ctx, daemon.Options{Once: *once, PollEvery: *poll, ConfigPath: *configPath})
}

func runApply(args []string) error {
	flags := flag.NewFlagSet("apply", flag.ExitOnError)
	phaseID := flags.String("phase", "", "phase id to apply")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *phaseID == "" {
		return fmt.Errorf("--phase is required")
	}
	cfg, paths, err := config.LoadOrCreateDefault()
	if err != nil {
		return err
	}
	results, err := daemon.ApplyPhaseByID(context.Background(), cfg, paths.Snapshots, *phaseID)
	if err != nil {
		return err
	}
	data, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(data))
	return nil
}

func runDoctor() error {
	fmt.Print(doctor.Format(doctor.Run(context.Background())))
	return nil
}

func runInstallUserService(args []string) error {
	flags := flag.NewFlagSet("install-user-service", flag.ExitOnError)
	enableNow := flags.Bool("now", true, "enable and start now")
	binary := flags.String("binary", "", "binary path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	path, err := systemd.InstallUserService(*binary, *enableNow)
	if err != nil {
		return err
	}
	fmt.Println(path)
	return nil
}

func runInstallPrivilegedSchedule() error {
	cfg, _, err := config.LoadOrCreateDefault()
	if err != nil {
		return err
	}
	filtered := privileged.FilterConfig(cfg)
	schedule := privileged.Schedule{
		Version: 1,
		UserUID: strconv.Itoa(os.Getuid()),
		Config:  filtered,
		Written: time.Now(),
	}
	data, err := json.Marshal(schedule)
	if err != nil {
		return err
	}
	helper, err := rootctlHelperPath()
	if err != nil {
		return err
	}
	cmd := exec.Command("pkexec", helper, "install-schedule", "--stdin", "--user", schedule.UserUID)
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runShowConfig() error {
	cfg, paths, err := config.LoadOrCreateDefault()
	if err != nil {
		return err
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	fmt.Printf("%s\n%s\n", paths.Config, data)
	return nil
}

func printUsage() {
	fmt.Println(`ThemeTime

Usage:
  themetime gui
  themetime daemon [--once] [--poll 15s]
  themetime apply --phase <id>
  themetime doctor
  themetime install-user-service [--now=true]
  themetime install-privileged-schedule
  themetime show-config
  themetime version`)
}

func guiCommand() (*exec.Cmd, error) {
	if explicit := os.Getenv("THEMETIME_GUI"); explicit != "" {
		return exec.Command(explicit), nil
	}
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		for _, name := range guiBinaryNames() {
			candidate := filepath.Join(dir, name)
			if _, err := os.Stat(candidate); err == nil {
				return exec.Command(candidate), nil
			}
		}
	}
	if _, err := os.Stat("cmd/themetime-wails"); err == nil {
		if _, err := exec.LookPath("go"); err == nil {
			return exec.Command("go", "run", "-tags", wailsBuildTags(), "./cmd/themetime-wails"), nil
		}
	}
	for _, name := range guiBinaryNames() {
		if path, err := exec.LookPath(name); err == nil {
			return exec.Command(path), nil
		}
	}
	return nil, fmt.Errorf("ThemeTime GUI was not found; build it with `make build`")
}

func guiBinaryNames() []string {
	return []string{"themetime-wails", "themetime-wails.exe"}
}

func wailsBuildTags() string {
	tags := "production,desktop"
	if _, err := exec.LookPath("pkg-config"); err == nil {
		if err := exec.Command("pkg-config", "--exists", "webkit2gtk-4.1").Run(); err == nil {
			tags += ",webkit2_41"
		}
	}
	return tags
}

func rootctlHelperPath() (string, error) {
	info, err := os.Stat(rootctlPolicyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("privileged helper is not installed at %s; run `sudo make install-root-assets`", rootctlPolicyPath)
		}
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory, expected the themetime-rootctl executable", rootctlPolicyPath)
	}
	if info.Mode().Perm()&0o111 == 0 {
		return "", fmt.Errorf("%s is not executable", rootctlPolicyPath)
	}
	return rootctlPolicyPath, nil
}
