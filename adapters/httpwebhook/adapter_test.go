package httpwebhook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

func TestIngestSendsAssertionThroughSDK(t *testing.T) {
	core := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/assertions" || r.Header.Get("X-Tenant-ID") != "acme" || r.Header.Get("Idempotency-Key") != "webhook-1" {
			t.Fatalf("request = %s tenant=%q idempotency=%q", r.URL.Path, r.Header.Get("X-Tenant-ID"), r.Header.Get("Idempotency-Key"))
		}
		var body ontovela.StateAssertionInput
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.StateKind != "observed" {
			t.Fatalf("state kind = %q", body.StateKind)
		}
		_ = json.NewEncoder(w).Encode(ontovela.StateAssertion{ID: "a1"})
	}))
	defer core.Close()

	client, err := ontovela.NewClient(core.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	request := IngestRequest{
		TenantID: "acme", IdempotencyKey: "webhook-1", SubjectID: "robot:WH-17", Property: "health",
		Value: json.RawMessage(`"ready"`), StateKind: "observed", EventTime: time.Now().UTC(),
		Source: "webhook:robot", EvidenceRef: "webhook:event/1",
	}
	claim, err := Ingest(context.Background(), client, request)
	if err != nil {
		t.Fatal(err)
	}
	if claim.ID != "a1" {
		t.Fatalf("claim = %#v", claim)
	}
}

func TestServeHTTPIngestsWebhookBody(t *testing.T) {
	core := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(ontovela.StateAssertion{ID: "a1"})
	}))
	defer core.Close()
	client, err := ontovela.NewClient(core.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	server := &Server{Client: client}
	body := []byte(`{"tenant_id":"acme","idempotency_key":"k","subject_id":"robot:WH-17","property":"health","value":"ready","state_kind":"observed","event_time":"2026-08-11T10:00:00Z","source":"webhook:robot","evidence_ref":"e1"}`)
	request := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(body))
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestIngestRejectsInvalidStateKind(t *testing.T) {
	client, err := ontovela.NewClient("http://localhost:1", "acme")
	if err != nil {
		t.Fatal(err)
	}
	request := IngestRequest{
		TenantID: "acme", IdempotencyKey: "k", SubjectID: "robot:WH-17", Property: "health",
		Value: json.RawMessage(`"ready"`), StateKind: "fact", EventTime: time.Now().UTC(),
		Source: "webhook:robot", EvidenceRef: "e1",
	}
	if _, err := Ingest(context.Background(), client, request); err == nil {
		t.Fatal("expected invalid state kind rejection")
	}
}
