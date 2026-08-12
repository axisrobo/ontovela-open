package can

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/axisrobo/ONTOVELA-open/adapters/base"
	"github.com/axisrobo/ONTOVELA-open/adapters/base/testutil"
	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

func TestRunIngestsMessage(t *testing.T) {
	core := testutil.NewFakeCore()
	defer core.Close()
	client, err := ontovela.NewClient(core.URL(), "acme")
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(base.Payload{TenantID: "acme", IdempotencyKey: "k1", SubjectID: "robot:WH-17", Property: "health", Value: json.RawMessage("\"ready\""), StateKind: "observed", EventTime: time.Now().UTC(), Source: "adapter:can", EvidenceRef: "can/1"})
	if err != nil {
		t.Fatal(err)
	}
	source := &MemorySource{Messages: []Message{{Body: payload}}}
	if err := Run(context.Background(), source, client); err != nil {
		t.Fatal(err)
	}
	if core.Count() != 1 {
		t.Fatalf("appends = %d", core.Count())
	}
}
