package kde

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Runner interface {
	LookPath(name string) (string, error)
	Run(ctx context.Context, name string, args ...string) (string, error)
}

type ExecRunner struct {
	Timeout time.Duration
}

func (r ExecRunner) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

func (r ExecRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	timeout := r.Timeout
	if timeout == 0 {
		timeout = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	text := strings.TrimSpace(out.String())
	if ctx.Err() != nil {
		return text, ctx.Err()
	}
	if err != nil {
		if text != "" {
			return text, fmt.Errorf("%s %v: %w: %s", name, args, err, text)
		}
		return text, fmt.Errorf("%s %v: %w", name, args, err)
	}
	return text, nil
}

func commandExists(r Runner, name string) bool {
	_, err := r.LookPath(name)
	return err == nil
}

func bestCommand(r Runner, names ...string) (string, bool) {
	for _, name := range names {
		if commandExists(r, name) {
			return name, true
		}
	}
	return "", false
}

func IsCommandMissing(err error) bool {
	var execErr *exec.Error
	return errors.As(err, &execErr)
}
