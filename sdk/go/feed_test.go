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
