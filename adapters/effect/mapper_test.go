package effect

import (
	"encoding/json"
	"testing"
	"time"
)

func TestToAssertionInputDefaultsToObserved(t *testing.T) {
	effect := EffectRecord{
		EffectID: "eff-1", TenantID: "acme", TwinID: "robot:WH-17", Property: "health",
		Value: json.RawMessage(`"ready"`), ObservedAt: time.Now().UTC(), Source: "kinetovela:robot-1", EvidenceRef: "effect/eff-1",
	}
	input, err := ToAssertionInput(effect)
	if err != nil {
		t.Fatal(err)
	}
	if input.StateKind != "observed" {
		t.Fatalf("state kind = %q", input.StateKind)
	}
	if input.EvidenceRef != "effect/eff-1" {
		t.Fatalf("evidence = %q", input.EvidenceRef)
	}
}

func TestToAssertionInputAllowsReportedOnly(t *testing.T) {
	effect := EffectRecord{
		EffectID: "eff-1", TenantID: "acme", TwinID: "robot:WH-17", Property: "task",
		Value: json.RawMessage(`"complete"`), ObservedAt: time.Now().UTC(), Source: "kinetovela:robot-1", EvidenceRef: "effect/eff-1", StateKind: "reported",
	}
	input, err := ToAssertionInput(effect)
	if err != nil {
		t.Fatal(err)
	}
	if input.StateKind != "reported" {
		t.Fatalf("state kind = %q", input.StateKind)
	}
}

func TestToAssertionInputRejectsSimulatedAndMissingFields(t *testing.T) {
	for _, kind := range []string{"simulated", "predicted", "inferred"} {
		effect := EffectRecord{
			EffectID: "eff-1", TenantID: "acme", TwinID: "robot:WH-17", Property: "health",
			Value: json.RawMessage(`"ready"`), ObservedAt: time.Now().UTC(), Source: "kinetovela:robot-1", EvidenceRef: "effect/eff-1", StateKind: kind,
		}
		if _, err := ToAssertionInput(effect); err == nil {
			t.Errorf("effect with state kind %q must be rejected", kind)
		}
	}
	if _, err := ToAssertionInput(EffectRecord{EffectID: "x", TenantID: "acme"}); err == nil {
		t.Fatal("incomplete effect must be rejected")
	}
}
