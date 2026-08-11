package ontovela

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAppendAssertionSendsTenantAndIdempotencyHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/assertions" || r.Header.Get("X-Tenant-ID") != "acme" || r.Header.Get("Idempotency-Key") != "event-1" {
			t.Fatalf("request = %s tenant=%q idempotency=%q", r.URL.Path, r.Header.Get("X-Tenant-ID"), r.Header.Get("Idempotency-Key"))
		}
		_ = json.NewEncoder(w).Encode(StateAssertion{ID: "assertion-1", TenantID: "acme", SystemTime: time.Now().UTC()})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	claim, err := client.AppendAssertion(context.Background(), StateAssertionInput{SubjectID: "robot:wh-17", Property: "health", Value: json.RawMessage(`"ready"`), StateKind: Observed, EventTime: time.Now().UTC(), Source: "sensor:health", EvidenceRef: "event:1"}, "event-1")
	if err != nil {
		t.Fatal(err)
	}
	if claim.ID != "assertion-1" {
		t.Fatalf("claim = %#v", claim)
	}
}

func TestResolveStateSendsBitemporalQuery(t *testing.T) {
	when := time.Date(2026, 8, 11, 10, 0, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("as_of") == "" || r.URL.Query().Get("as_known") == "" {
			t.Fatalf("query = %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(ResolvedState{Status: "resolved", Property: "health", Value: json.RawMessage(`"ready"`)})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	state, err := client.ResolveState(context.Background(), "robot:wh-17", "health", TemporalQuery{AsOf: &when, AsKnown: &when})
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != "resolved" {
		t.Fatalf("state = %#v", state)
	}
}

func TestCreateRealityViewPostsPurposeAndRequiredState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload RealityViewRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if r.URL.Path != "/v1/reality-views" || payload.Purpose != "dispatch" || len(payload.RequiredState) != 1 || payload.RequiredState[0].Property != "health" {
			t.Fatalf("request path=%s payload=%#v", r.URL.Path, payload)
		}
		_ = json.NewEncoder(w).Encode(RealityView{Status: "ready", Items: []RealityViewItem{{Property: "health", Acceptable: true}}})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	view, err := client.CreateRealityView(context.Background(), RealityViewRequest{TwinID: "robot:wh-17", Purpose: "dispatch", RequiredState: []RequiredState{{Property: "health", MaxAgeSeconds: 60}}}, TemporalQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if view.Status != "ready" || len(view.Items) != 1 || !view.Items[0].Acceptable {
		t.Fatalf("view = %#v", view)
	}
}

func TestCommitSubscriptionOffsetSendsConsumerAndOffset(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/subscriptions/planner-1/commit" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var payload struct{ Offset int64 }
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.Offset != 5 {
			t.Fatalf("offset = %d", payload.Offset)
		}
		_ = json.NewEncoder(w).Encode(SubscriptionOffset{ConsumerID: "planner-1", CommittedOffset: 5})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	cursor, err := client.CommitSubscriptionOffset(context.Background(), "planner-1", 5)
	if err != nil || cursor.CommittedOffset != 5 {
		t.Fatalf("cursor=%#v err=%v", cursor, err)
	}
}

func TestListConflictsFiltersByStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/conflicts" || r.URL.Query().Get("status") != "open" {
			t.Fatalf("path=%s query=%s", r.URL.Path, r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"conflicts": []ConflictRecord{{SubjectID: "bin:A-01", Property: "available", Status: "open"}}})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	records, err := client.ListConflicts(context.Background(), "open", 100)
	if err != nil || len(records) != 1 || records[0].Status != "open" {
		t.Fatalf("records=%#v err=%v", records, err)
	}
}

func TestListTwinTypesParsesPacks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/twin-types" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"twin_types": []TwinType{{TypeRef: "robot", Description: "Robot twin", Properties: []string{"location"}}}})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	types, err := client.ListTwinTypes(context.Background())
	if err != nil || len(types) != 1 || types[0].TypeRef != "robot" {
		t.Fatalf("types=%#v err=%v", types, err)
	}
}

func TestComputeImpactSendsDepthAndParsesPaths(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/twins/robot:wh-17/impact" || r.URL.Query().Get("max_depth") != "3" {
			t.Fatalf("path=%s query=%s", r.URL.Path, r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"impact_paths": []ImpactPath{{SubjectID: "robot:wh-17", TargetID: "zone:charging", Predicate: "located_in", Depth: 1}}})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	paths, err := client.ComputeImpact(context.Background(), "robot:wh-17", 3, "located_in", TemporalQuery{})
	if err != nil || len(paths) != 1 || paths[0].TargetID != "zone:charging" {
		t.Fatalf("paths=%#v err=%v", paths, err)
	}
}

func TestListAssertionsAndAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/twins/robot:wh-17/assertions":
			if r.URL.Query().Get("property") != "health" {
				t.Fatalf("query = %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"assertions": []StateAssertion{{ID: "assertion-1"}}})
		case "/v1/twins/missing":
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		default:
			t.Fatalf("path = %s", r.URL.Path)
		}
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	assertions, err := client.ListAssertions(context.Background(), "robot:wh-17", "health", TemporalQuery{})
	if err != nil || len(assertions) != 1 || assertions[0].ID != "assertion-1" {
		t.Fatalf("assertions=%#v err=%v", assertions, err)
	}
	_, err = client.GetTwin(context.Background(), "missing")
	apiError, ok := err.(*APIError)
	if !ok || apiError.StatusCode != http.StatusNotFound || apiError.Message != "not found" {
		t.Fatalf("error = %#v", err)
	}
}
