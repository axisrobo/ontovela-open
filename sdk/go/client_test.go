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
