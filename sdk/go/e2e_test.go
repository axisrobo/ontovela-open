package ontovela

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// e2eCore is a minimal stateful fake of the ONTOVELA core for SDK end-to-end
// tests. It is not a product implementation.
type e2eCore struct {
	mu         sync.Mutex
	assertions []StateAssertion
	events     []ChangeEvent
	offset     int64
}

func (e *e2eCore) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case r.Method == http.MethodPost && path == "/v1/twins":
			_ = json.NewEncoder(w).Encode(Twin{TwinInput: TwinInput{ID: "robot:WH-17", TypeRef: "robot"}, TenantID: "acme"})
		case r.Method == http.MethodPost && path == "/v1/source-bindings":
			var input SourceBindingInput
			_ = json.NewDecoder(r.Body).Decode(&input)
			_ = json.NewEncoder(w).Encode(SourceBinding{SourceBindingInput: input, TenantID: "acme"})
		case r.Method == http.MethodPost && path == "/v1/assertions":
			var input StateAssertionInput
			_ = json.NewDecoder(r.Body).Decode(&input)
			e.mu.Lock()
			e.offset++
			claim := StateAssertion{StateAssertionInput: input, ID: "a1", TenantID: "acme", SystemTime: time.Now().UTC()}
			e.assertions = append(e.assertions, claim)
			e.events = append(e.events, ChangeEvent{Offset: e.offset, Kind: "state_assertion.appended", SubjectID: input.SubjectID})
			e.mu.Unlock()
			_ = json.NewEncoder(w).Encode(claim)
		case r.Method == http.MethodGet && strings.HasPrefix(path, "/v1/twins/") && strings.HasSuffix(path, "/state/health"):
			_ = json.NewEncoder(w).Encode(ResolvedState{Status: "resolved", Property: "health", Value: json.RawMessage(`"ready"`), Freshness: "fresh"})
		case r.Method == http.MethodPost && strings.HasPrefix(path, "/v1/twins/") && strings.HasSuffix(path, "/snapshots"):
			_ = json.NewEncoder(w).Encode(Snapshot{ID: "snap-1", TenantID: "acme", SubjectID: "robot:WH-17", Digest: "d1"})
		case r.Method == http.MethodGet && strings.Contains(path, "/verify"):
			_ = json.NewEncoder(w).Encode(map[string]any{"valid": true})
		case r.Method == http.MethodGet && path == "/v1/changes":
			e.mu.Lock()
			defer e.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"events": e.events})
		default:
			http.NotFound(w, r)
		}
	})
}

func TestGoSDKEndToEndWarehouseFlow(t *testing.T) {
	core := &e2eCore{}
	server := httptest.NewServer(core.handler())
	defer server.Close()

	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := client.CreateTwin(ctx, TwinInput{ID: "robot:WH-17", TypeRef: "robot"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateSourceBinding(ctx, SourceBindingInput{ID: "b1", Source: "sensor:health", Property: "health", AuthorityRank: 1}); err != nil {
		t.Fatal(err)
	}
	claim, err := client.AppendAssertion(ctx, StateAssertionInput{SubjectID: "robot:WH-17", Property: "health", Value: json.RawMessage(`"ready"`), StateKind: Observed, EventTime: time.Now().UTC(), Source: "sensor:health", EvidenceRef: "e1"}, "k1")
	if err != nil || claim.ID != "a1" {
		t.Fatalf("claim=%#v err=%v", claim, err)
	}
	state, err := client.ResolveState(ctx, "robot:WH-17", "health", TemporalQuery{})
	if err != nil || state.Status != "resolved" {
		t.Fatalf("state=%#v err=%v", state, err)
	}
	snapshot, err := client.CreateSnapshot(ctx, "robot:WH-17", TemporalQuery{})
	if err != nil || snapshot.ID != "snap-1" {
		t.Fatalf("snapshot=%#v err=%v", snapshot, err)
	}
	valid, err := client.VerifySnapshot(ctx, snapshot.ID)
	if err != nil || !valid {
		t.Fatalf("verify err=%v", err)
	}
	events, err := client.ListChanges(ctx, 0, 100, ChangeFilter{})
	if err != nil || len(events) != 1 {
		t.Fatalf("events=%#v err=%v", events, err)
	}
}
