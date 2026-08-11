package mqtt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

func TestRunIngestsMQTTMessage(t *testing.T) {
	core := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(ontovela.StateAssertion{ID: "a1"})
	}))
	defer core.Close()
	client, err := ontovela.NewClient(core.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(Payload{
		TenantID: "acme", IdempotencyKey: "k1", SubjectID: "robot:WH-17", Property: "health",
		Value: json.RawMessage(`"ready"`), StateKind: "observed", EventTime: time.Now().UTC(), Source: "mqtt:robot", EvidenceRef: "mqtt/1",
	})
	source := &MemorySource{Messages: []Message{{Topic: "ontovela/state/robot/WH-17", Payload: payload}}}
	if err := Run(context.Background(), source, []string{"ontovela/state/+/+"}, client); err != nil {
		t.Fatal(err)
	}
}

func TestRunRejectsCrossTenant(t *testing.T) {
	client, err := ontovela.NewClient("http://localhost:1", "acme")
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(Payload{TenantID: "other", IdempotencyKey: "k", SubjectID: "x", Property: "y", Value: json.RawMessage(`1`), StateKind: "observed", EventTime: time.Now().UTC(), Source: "mqtt", EvidenceRef: "e"})
	source := &MemorySource{Messages: []Message{{Topic: "t", Payload: payload}}}
	if err := Run(context.Background(), source, nil, client); err == nil {
		t.Fatal("expected cross-tenant rejection")
	}
}

func TestTopicFromID(t *testing.T) {
	if got := TopicFromID("robot:WH-17"); got != "ontovela/state/robot/WH-17" {
		t.Fatalf("topic = %q", got)
	}
}
