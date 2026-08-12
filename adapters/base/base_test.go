package base

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

func TestValidateAndAppend(t *testing.T) {
	payload := Payload{
		TenantID: "acme", IdempotencyKey: "k1", SubjectID: "robot:WH-17", Property: "health",
		Value: json.RawMessage(`"ready"`), StateKind: "observed", EventTime: time.Now().UTC(),
		Source: "sensor:health", EvidenceRef: "e1",
	}
	if err := payload.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"", "fact"} {
		bad := payload
		bad.StateKind = kind
		if err := bad.Validate(); err == nil {
			t.Fatalf("state kind %q must be rejected", kind)
		}
	}
}

func TestAppendTenantMismatch(t *testing.T) {
	client, err := ontovela.NewClient("http://localhost:1", "acme")
	if err != nil {
		t.Fatal(err)
	}
	payload := Payload{TenantID: "other", IdempotencyKey: "k", SubjectID: "x", Property: "y", Value: json.RawMessage(`1`), StateKind: "observed", EventTime: time.Now().UTC(), Source: "s", EvidenceRef: "e"}
	if err := Append(context.Background(), client, payload); err == nil {
		t.Fatal("tenant mismatch must be rejected")
	}
}
