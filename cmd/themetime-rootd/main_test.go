package main

import (
	"context"
	"errors"
	"testing"
)

func TestRunWithArgsOnceReturnsTickError(t *testing.T) {
	want := errors.New("tick failed")
	err := runWithArgs([]string{"--once"}, func(context.Context) error {
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}
