package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/RedUndercover/themetime/internal/config"
	"github.com/RedUndercover/themetime/internal/privileged"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 || os.Args[1] != "install-schedule" {
		return fmt.Errorf("usage: themetime-rootctl install-schedule --stdin --user <uid>")
	}
	flags := flag.NewFlagSet("install-schedule", flag.ExitOnError)
	fromStdin := flags.Bool("stdin", false, "read schedule JSON from stdin")
	user := flags.String("user", "", "owning user uid")
	if err := flags.Parse(os.Args[2:]); err != nil {
		return err
	}
	if !*fromStdin {
		return fmt.Errorf("--stdin is required")
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	var schedule privileged.Schedule
	if err := json.Unmarshal(data, &schedule); err != nil {
		return err
	}
	if *user != "" && schedule.UserUID != *user {
		return fmt.Errorf("schedule user %q does not match requested user %q", schedule.UserUID, *user)
	}
	if schedule.Written.IsZero() {
		schedule.Written = time.Now()
	}
	return privileged.InstallSchedule(config.RootPaths().Config, schedule)
}
