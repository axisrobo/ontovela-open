package stream

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

func mustTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return parsed
}

func TestRunIngestsStreamMessages(t *testing.T) {
	core := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(ontovela.StateAssertion{ID: "a1"})
	}))
	defer core.Close()
	client, err := ontovela.NewClient(core.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(IngestRequest{
		TenantID: "acme", IdempotencyKey: "k1", SubjectID: "robot:WH-17", Property: "health",
		Value: json.RawMessage(`"ready"`), StateKind: "observed", EventTime: mustTime("2026-08-11T10:00:00Z"),
		Source: "kafka:robot", EvidenceRef: "event/1",
	})
	source := &MemorySource{Messages: []Message{{Body: body, Topic: "robot-state", Partition: 0, Offset: 1}}}
	if err := Run(context.Background(), source, client); err != nil {
		t.Fatal(err)
	}
}

func TestRunRejectsCrossTenantMessage(t *testing.T) {
	client, err := ontovela.NewClient("http://localhost:1", "acme")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(IngestRequest{
		TenantID: "other", IdempotencyKey: "k1", SubjectID: "robot:WH-17", Property: "health",
		Value: json.RawMessage(`"ready"`), StateKind: "observed", EventTime: mustTime("2026-08-11T10:00:00Z"),
		Source: "kafka:robot", EvidenceRef: "event/1",
	})
	source := &MemorySource{Messages: []Message{{Body: body}}}
	if err := Run(context.Background(), source, client); err == nil {
		t.Fatal("expected cross-tenant rejection")
	}
}

func TestRunRejectsInvalidStateKind(t *testing.T) {
	client, err := ontovela.NewClient("http://localhost:1", "acme")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(IngestRequest{
		TenantID: "acme", IdempotencyKey: "k1", SubjectID: "robot:WH-17", Property: "health",
		Value: json.RawMessage(`"ready"`), StateKind: "fact", EventTime: mustTime("2026-08-11T10:00:00Z"),
		Source: "kafka:robot", EvidenceRef: "event/1",
	})
	source := &MemorySource{Messages: []Message{{Body: body}}}
	if err := Run(context.Background(), source, client); err == nil {
		t.Fatal("expected invalid state kind rejection")
	}
}
