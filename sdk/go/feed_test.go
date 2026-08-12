package ontovela

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestConsumeChangesFetchesAndCommits(t *testing.T) {
	var commits []int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/subscriptions/consumer-1" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(SubscriptionOffset{CommittedOffset: 0})
		case r.URL.Path == "/v1/changes":
			_ = json.NewEncoder(w).Encode(map[string]any{"events": []ChangeEvent{{Offset: 1}, {Offset: 2}}})
		case r.URL.Path == "/v1/subscriptions/consumer-1/commit" && r.Method == http.MethodPost:
			var payload struct{ Offset int64 }
			_ = json.NewDecoder(r.Body).Decode(&payload)
			commits = append(commits, payload.Offset)
			_ = json.NewEncoder(w).Encode(SubscriptionOffset{CommittedOffset: payload.Offset})
		default:
			t.Fatalf("path=%s method=%s", r.URL.Path, r.Method)
		}
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	events, err := client.ConsumeChanges(context.Background(), "consumer-1", 100, ChangeFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || len(commits) != 1 || commits[0] != 2 {
		t.Fatalf("events=%#v commits=%v", events, commits)
	}
}

func TestStreamChangesDrainsUntilDone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/audit/changes/stream" || r.URL.Query().Get("after") != "0" {
			t.Fatalf("path=%s query=%s", r.URL.Path, r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		_, _ = w.Write([]byte(`{"after":2,"count":2,"events":[{"offset":1},{"offset":2}]}` + "\n"))
		_, _ = w.Write([]byte(`{"after":2,"count":0,"events":[]}` + "\n"))
		_, _ = w.Write([]byte(`{"done":true,"after":2}` + "\n"))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	var pages [][]ChangeEvent
	lastAfter, err := client.StreamAuditChanges(context.Background(), 0, func(page []ChangeEvent, nextAfter int64) bool {
		pages = append(pages, page)
		return true
	})
	if err != nil {
		t.Fatal(err)
	}
	if lastAfter != 2 || len(pages) != 1 || len(pages[0]) != 2 {
		t.Fatalf("lastAfter=%d pages=%d first=%d", lastAfter, len(pages), len(pages[0]))
	}
}
