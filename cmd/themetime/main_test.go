package main

import (
	"slices"
	"strings"
	"testing"
)

func TestGUIBinaryNamesAreWailsOnly(t *testing.T) {
	names := guiBinaryNames()
	if !slices.Contains(names, "themetime-wails") {
		t.Fatalf("GUI candidates = %v, want themetime-wails", names)
	}
	for _, name := range names {
		if strings.Contains(name, "themetime-gui") {
			t.Fatalf("GUI candidates still include legacy Fyne binary: %v", names)
		}
	}
}
