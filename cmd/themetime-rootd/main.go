package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/themetime/themetime/internal/config"
	"github.com/themetime/themetime/internal/kde"
	"github.com/themetime/themetime/internal/privileged"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	return runWithArgs(os.Args[1:], tick)
}

func runWithArgs(args []string, tickFn func(context.Context) error) error {
	flags := flag.NewFlagSet("themetime-rootd", flag.ExitOnError)
	once := flags.Bool("once", false, "apply once and exit")
	poll := flags.Duration("poll", time.Minute, "poll interval")
	if err := flags.Parse(args); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	for {
		err := tickFn(ctx)
		if *once {
			return err
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(*poll):
		}
	}
}

func tick(ctx context.Context) error {
	paths := config.RootPaths()
	schedule, err := privileged.LoadSchedule(paths.Config)
	if err != nil {
		return err
	}
	results, err := privileged.ApplyDueStateful(ctx, kde.ExecRunner{}, schedule, paths.State, time.Now())
	if err != nil {
		return err
	}
	if len(results) > 0 {
		data, _ := json.Marshal(results)
		fmt.Println(string(data))
	}
	return nil
}
