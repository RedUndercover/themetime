package model

import (
	"slices"
	"strings"
	"testing"
)

func TestTriggerDefinitionsAreOrderedAndUnique(t *testing.T) {
	definitions := TriggerDefinitions()
	if len(definitions) != 10 {
		t.Fatalf("definitions = %d, want 10", len(definitions))
	}
	if definitions[0].Kind != TriggerClock {
		t.Fatalf("first trigger = %q, want clock", definitions[0].Kind)
	}
	seen := map[TriggerKind]bool{}
	for _, definition := range definitions {
		if seen[definition.Kind] {
			t.Fatalf("duplicate trigger %q", definition.Kind)
		}
		seen[definition.Kind] = true
		if definition.Label == "" || definition.ShortLabel == "" {
			t.Fatalf("trigger %q has an empty label", definition.Kind)
		}
	}
	wantSolar := make([]TriggerKind, 0, len(definitions)-1)
	for _, definition := range definitions[1:] {
		wantSolar = append(wantSolar, definition.Kind)
	}
	if !slices.Equal(SolarTriggerKinds(), wantSolar) {
		t.Fatalf("solar kinds = %v, want %v", SolarTriggerKinds(), wantSolar)
	}

	copy := TriggerDefinitions()
	copy[0].Label = "changed"
	if TriggerDefinitions()[0].Label != "Clock" {
		t.Fatal("TriggerDefinitions returned mutable shared state")
	}
}

func TestTriggerLabelsComeFromDefinitions(t *testing.T) {
	for _, definition := range TriggerDefinitions() {
		if got := SolarTriggerLabel(definition.Kind); got != definition.Label {
			t.Fatalf("label for %q = %q, want %q", definition.Kind, got, definition.Label)
		}
		if got := SolarTriggerShortLabel(definition.Kind); got != definition.ShortLabel {
			t.Fatalf("short label for %q = %q, want %q", definition.Kind, got, definition.ShortLabel)
		}
	}
}

func TestActionTypesAreUniqueLabeledAndValidated(t *testing.T) {
	types := AllActionTypes()
	seen := map[ActionType]bool{}
	for _, actionType := range types {
		if seen[actionType] {
			t.Fatalf("duplicate action type %q", actionType)
		}
		seen[actionType] = true
		if label := (Action{Type: actionType}).Label(); label == "" || label == string(actionType) {
			t.Fatalf("action type %q has no friendly label", actionType)
		}
		if priority := ActionPriority(actionType); priority < 0 || priority > 4 {
			t.Fatalf("action type %q has invalid priority %d", actionType, priority)
		}
	}
	if len(types) != 13 {
		t.Fatalf("action types = %d, want 13", len(types))
	}
	for _, actionType := range []ActionType{ActionSDDMTheme, ActionPlymouthTheme} {
		if !(Action{Type: actionType}).IsPrivileged() {
			t.Fatalf("action type %q should be privileged", actionType)
		}
	}
	if err := (Action{Type: ActionType("unknown"), Value: "value"}).Validate(); err == nil {
		t.Fatal("unknown action type passed validation")
	}
}

func TestConfigValidationCollectsStructuralErrors(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Location.Latitude = 91
	cfg.Phases[1].ID = cfg.Phases[0].ID
	cfg.Phases[0].Start = Trigger{Kind: TriggerKind("unknown")}
	cfg.Phases[1].Actions = []Action{{Type: ActionCustomCommand}}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected validation errors")
	}
	for _, fragment := range []string{"latitude", "duplicated", "unknown trigger", "custom command"} {
		if !strings.Contains(err.Error(), fragment) {
			t.Fatalf("validation error %q does not contain %q", err, fragment)
		}
	}
}
