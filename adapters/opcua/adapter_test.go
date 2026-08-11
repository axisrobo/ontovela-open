package opcua

import (
	"encoding/json"
	"testing"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

func TestToAssertionInputMapsNodeRead(t *testing.T) {
	node := NodeValue{
		TenantID: "acme", IdempotencyKey: "k1", TwinID: "pump:P-1", Property: "speed",
		Value: json.RawMessage(`1200`), Quality: "good", ObservedAt: time.Now().UTC(), Source: "opcua:line-1", EvidenceRef: "opcua/1",
	}
	input, err := ToAssertionInput(node)
	if err != nil {
		t.Fatal(err)
	}
	if input.StateKind != ontovela.Observed || input.SubjectID != "pump:P-1" {
		t.Fatalf("input = %#v", input)
	}
}

func TestToAssertionInputRejectsBadQuality(t *testing.T) {
	node := NodeValue{
		TenantID: "acme", IdempotencyKey: "k1", TwinID: "pump:P-1", Property: "speed",
		Value: json.RawMessage(`1200`), Quality: "bad", ObservedAt: time.Now().UTC(), Source: "opcua:line-1", EvidenceRef: "opcua/1",
	}
	if _, err := ToAssertionInput(node); err == nil {
		t.Fatal("bad quality must be rejected")
	}
}
