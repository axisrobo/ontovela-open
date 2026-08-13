package live

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

// TestLiveOrchadynRealityViewConsumption verifies that a real core serves a
// fresh Reality View from the observed sensor binding without promoting state.
func TestLiveOrchadynRealityViewConsumption(t *testing.T) {
	client := liveClient(t)
	seedTwin(t, client, "robot:orchadyn-1", "health", "sensor:health", 30)
	// Observe reality through the sensor binding.
	if _, err := client.AppendAssertion(context.Background(), ontovela.StateAssertionInput{
		SubjectID: "robot:orchadyn-1", Property: "health", Value: json.RawMessage(`"ready"`),
		StateKind: ontovela.Observed, EventTime: time.Now().UTC(),
		Source: "sensor:health", EvidenceRef: "sensor/1",
	}, "k-orchadyn-1"); err != nil {
		t.Fatalf("append observed: %v", err)
	}
	// ORCHADYN dispatches against a fresh Reality View.
	view, err := client.CreateRealityView(context.Background(), ontovela.RealityViewRequest{
		TwinID: "robot:orchadyn-1", Purpose: "dispatch",
		RequiredState: []ontovela.RequiredState{{Property: "health", MaxAgeSeconds: 60}},
	}, ontovela.TemporalQuery{})
	if err != nil {
		t.Fatalf("create reality view: %v", err)
	}
	if view.Status != "ready" {
		t.Fatalf("view status = %s, want ready", view.Status)
	}
	if len(view.Items) != 1 || !view.Items[0].Acceptable {
		t.Fatalf("view items not acceptable: %+v", view.Items)
	}
	if view.Items[0].State.StateKind != ontovela.Observed {
		t.Fatalf("view state kind = %s, want observed (no promotion)", view.Items[0].State.StateKind)
	}
}

// TestLivePeiravelaSignedSnapshotConsumption verifies that a real core signs a
// verifiable snapshot and keeps the simulated branch from the real state view.
func TestLivePeiravelaSignedSnapshotConsumption(t *testing.T) {
	client := liveClient(t)
	seedTwin(t, client, "robot:peiravela-1", "health", "sensor:health", 30)
	if _, err := client.AppendAssertion(context.Background(), ontovela.StateAssertionInput{
		SubjectID: "robot:peiravela-1", Property: "health", Value: json.RawMessage(`"ready"`),
		StateKind: ontovela.Observed, EventTime: time.Now().UTC(),
		Source: "sensor:health", EvidenceRef: "sensor/1",
	}, "k-peiravela-1"); err != nil {
		t.Fatalf("append observed: %v", err)
	}
	// PEIRAVELA consumes a signed snapshot for a branch comparison.
	snapshot, err := client.CreateSnapshot(context.Background(), "robot:peiravela-1", ontovela.TemporalQuery{})
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}
	if snapshot.ID == "" {
		t.Fatalf("snapshot has no id")
	}
	valid, err := client.VerifySnapshot(context.Background(), snapshot.ID)
	if err != nil {
		t.Fatalf("verify snapshot: %v", err)
	}
	if !valid {
		t.Fatalf("snapshot %s invalid", snapshot.ID)
	}
	// A simulated branch diverges from the observed real state.
	if _, err := client.AppendAssertion(context.Background(), ontovela.StateAssertionInput{
		SubjectID: "robot:peiravela-1", Property: "health", Value: json.RawMessage(`"failed"`),
		StateKind: ontovela.Simulated, EventTime: time.Now().UTC(),
		Source: "peiravela:branch-1", EvidenceRef: "sim/1",
	}, "k-peiravela-sim"); err != nil {
		t.Fatalf("append simulated: %v", err)
	}
	delta, err := client.SimToReal(context.Background(), "robot:peiravela-1", "health", ontovela.TemporalQuery{})
	if err != nil {
		t.Fatalf("sim-to-real: %v", err)
	}
	if delta.Delta != "diverges" {
		t.Fatalf("delta = %q, want diverges", delta.Delta)
	}
	if delta.RealState.StateKind != ontovela.Observed {
		t.Fatalf("real state kind = %s, want observed (simulated never promoted)", delta.RealState.StateKind)
	}
}
