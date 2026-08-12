package edge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

func item(key string) Item {
	return Item{
		TenantID: "acme", IdempotencyKey: key, SubjectID: "robot:WH-17", Property: "health",
		Value: json.RawMessage(`"ready"`), StateKind: "observed", EventTime: mustTime("2026-08-11T10:00:00Z"),
		Source: "edge:robot", EvidenceRef: "edge/event/" + key,
	}
}

func TestFlushReplaysInOrderAndClears(t *testing.T) {
	var calls []string
	core := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body ontovela.StateAssertionInput
		_ = json.NewDecoder(r.Body).Decode(&body)
		calls = append(calls, body.EvidenceRef)
		_ = json.NewEncoder(w).Encode(ontovela.StateAssertion{ID: "a"})
	}))
	defer core.Close()
	client, err := ontovela.NewClient(core.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	spool := New()
	spool.Append(item("k1"))
	spool.Append(item("k2"))
	if err := spool.Flush(context.Background(), client); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 || calls[0] != "edge/event/k1" || calls[1] != "edge/event/k2" {
		t.Fatalf("calls = %#v", calls)
	}
	if spool.Len() != 0 {
		t.Fatalf("spool not cleared: %d", spool.Len())
	}
}

func TestFailedFlushRetainsItems(t *testing.T) {
	var fail atomic.Bool
	fail.Store(true)
	core := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fail.Load() {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(ontovela.StateAssertion{ID: "a"})
	}))
	defer core.Close()
	client, err := ontovela.NewClient(core.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	spool := New()
	spool.Append(item("k1"))
	if err := spool.Flush(context.Background(), client); err == nil {
		t.Fatal("expected flush failure")
	}
	if spool.Len() != 1 {
		t.Fatalf("failed item must be retained, len = %d", spool.Len())
	}
	// Once the source recovers, the buffered item replays.
	fail.Store(false)
	if err := spool.Flush(context.Background(), client); err != nil {
		t.Fatal(err)
	}
	if spool.Len() != 0 {
		t.Fatalf("spool not cleared after recovery, len = %d", spool.Len())
	}
}

func TestFileSpoolSurvivesReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "spool.jsonl")
	spool, err := NewFileSpool(path)
	if err != nil {
		t.Fatal(err)
	}
	first := spool.Append(item("k1"))
	reloaded, err := NewFileSpool(path)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.Len() != 1 || reloaded.items[0].Sequence != first {
		t.Fatalf("reloaded spool = %#v", reloaded.items)
	}
}

func mustTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return parsed
}
