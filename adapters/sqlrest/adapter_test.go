package sqlrest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

type memorySource struct {
	batch1 []Row
	batch2 []Row
	calls  int
}

func (m *memorySource) Fetch(_ context.Context, _ Cursor) ([]Row, Cursor, error) {
	m.calls++
	if m.calls == 1 {
		return m.batch1, "cursor-2", nil
	}
	return m.batch2, "cursor-3", nil
}

func row(key, idempotency string) Row {
	return Row{
		Cursor: Cursor(key), TenantID: "acme", IdempotencyKey: idempotency,
		SubjectID: "robot:WH-17", Property: "health", Value: json.RawMessage(`"ready"`),
		StateKind: "observed", EventTime: mustTime("2026-08-11T10:00:00Z"),
		Source: "sql:health", EvidenceRef: "row/" + key,
	}
}

func TestPollOnceAdvancesCursorAfterSuccess(t *testing.T) {
	core := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(ontovela.StateAssertion{ID: "a1"})
	}))
	defer core.Close()
	client, err := ontovela.NewClient(core.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	source := &memorySource{batch1: []Row{row("1", "k1")}, batch2: []Row{row("2", "k2")}}
	poller := &Poller{Source: source, Client: client}
	next, err := poller.PollOnce(context.Background(), "cursor-1")
	if err != nil {
		t.Fatal(err)
	}
	if next != "cursor-2" {
		t.Fatalf("cursor = %q", next)
	}
	// A failed batch must not advance the cursor.
	source2 := &memorySource{batch1: []Row{{Cursor: "bad", TenantID: "other", IdempotencyKey: "k", SubjectID: "x", Property: "y", Value: json.RawMessage(`1`), StateKind: "observed", EventTime: mustTime("2026-08-11T10:00:00Z"), Source: "sql", EvidenceRef: "e"}}}
	poller2 := &Poller{Source: source2, Client: client}
	next2, err := poller2.PollOnce(context.Background(), "cursor-1")
	if err == nil || next2 != "cursor-1" {
		t.Fatalf("failed batch must not advance: next=%q err=%v", next2, err)
	}
}

func mustTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return parsed
}
