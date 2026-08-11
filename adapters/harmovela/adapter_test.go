package harmovela

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func envelope(overrides map[string]any) map[string]any {
	base := map[string]any{
		"spec_version": "0.2",
		"id":           "event-1",
		"type":         "sensor.observation",
		"source":       "sensor:uwb-3",
		"created_at":   "2026-08-11T10:00:00Z",
		"payload":      map[string]any{"value": 12.44},
	}
	for key, value := range overrides {
		base[key] = value
	}
	return base
}

func TestFetchAndValidateAcceptsValidEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/events/event-1" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(envelope(nil))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	record, err := client.FetchAndValidate(context.Background(), "harmovela:event/event-1")
	if err != nil {
		t.Fatal(err)
	}
	if record.ID != "event-1" || record.Type != "sensor.observation" || record.CreatedAt.IsZero() {
		t.Fatalf("record = %#v", record)
	}
}

func TestFetchAndValidateRejectsInvalidEnvelope(t *testing.T) {
	cases := map[string]map[string]any{
		"wrong version":   {"spec_version": "0.1"},
		"missing source":  {"source": ""},
		"invalid created": {"created_at": "not-a-time"},
		"missing payload": nil,
	}
	for name, overrides := range cases {
		base := envelope(overrides)
		if overrides == nil {
			delete(base, "payload")
		}
		scoped := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(base)
		}))
		client, err := NewClient(scoped.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.FetchAndValidate(context.Background(), "harmovela:event/event-1"); err == nil {
			t.Errorf("%s: expected validation failure", name)
		}
		scoped.Close()
	}
}

func TestParseEvidenceRefRejectsUnsafeRefs(t *testing.T) {
	for _, ref := range []string{"", "http://x/event/1", "harmovela:event/", "harmovela:event/a b"} {
		if _, err := ParseEvidenceRef(ref); err == nil {
			t.Errorf("ParseEvidenceRef(%q) unexpectedly succeeded", ref)
		}
	}
}

func TestFetchAndValidateReportsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	client, err := NewClient(server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.FetchAndValidate(context.Background(), "harmovela:event/missing"); err == nil {
		t.Fatal("expected not-found error")
	}
}
