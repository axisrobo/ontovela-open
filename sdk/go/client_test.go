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

func TestSimToRealParsesDelta(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/twins/robot:wh-17/sim-to-real/health" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(SimToRealDelta{TwinID: "robot:wh-17", Property: "health", Delta: "diverges"})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	delta, err := client.SimToReal(context.Background(), "robot:wh-17", "health", TemporalQuery{})
	if err != nil || delta.Delta != "diverges" {
		t.Fatalf("delta=%#v err=%v", delta, err)
	}
}

func TestComputeCausalAnalyticsParsesCounts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/twins/robot:wh-17/causal/analytics" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(CausalAnalytics{TwinID: "robot:wh-17", FanOut: 3, FanIn: 1, TopTargets: map[string]int{"bin:A-01": 2}})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	analytics, err := client.ComputeCausalAnalytics(context.Background(), "robot:wh-17")
	if err != nil || analytics.FanOut != 3 || analytics.TopTargets["bin:A-01"] != 2 {
		t.Fatalf("analytics=%#v err=%v", analytics, err)
	}
}

func TestComputeCausalLineageParsesLinks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/twins/robot:wh-17/causal" || r.URL.Query().Get("max_depth") != "5" {
			t.Fatalf("path=%s query=%s", r.URL.Path, r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"causal_links": []CausalLink{{SubjectID: "robot:wh-17", TargetID: "bin:A-01", Mechanism: "motor_failure", Depth: 1}}})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	links, err := client.ComputeCausalLineage(context.Background(), "robot:wh-17", 5, TemporalQuery{})
	if err != nil || len(links) != 1 || links[0].Mechanism != "motor_failure" {
		t.Fatalf("links=%#v err=%v", links, err)
	}
}

func TestListSnapshotsParsesOrderedList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/twins/robot:wh-17/snapshots" || r.URL.Query().Get("limit") != "10" {
			t.Fatalf("path=%s query=%s", r.URL.Path, r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"snapshots": []Snapshot{{ID: "snap-1"}, {ID: "snap-2"}}})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	snapshots, err := client.ListSnapshots(context.Background(), "robot:wh-17", 10)
	if err != nil || len(snapshots) != 2 || snapshots[0].ID != "snap-1" {
		t.Fatalf("snapshots=%#v err=%v", snapshots, err)
	}
}

func TestReportHeartbeatPostsSource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/heartbeats" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var payload struct{ Source string }
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if payload.Source != "sensor:health" {
			t.Fatalf("source = %q", payload.Source)
		}
		_ = json.NewEncoder(w).Encode(SourceHeartbeat{TenantID: "acme", Source: "sensor:health"})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	heartbeat, err := client.ReportHeartbeat(context.Background(), "sensor:health")
	if err != nil || heartbeat.Source != "sensor:health" {
		t.Fatalf("heartbeat=%#v err=%v", heartbeat, err)
	}
}

func TestRetryWithBackoffOnTransientErrors(t *testing.T) {
	var attempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_ = json.NewEncoder(w).Encode(Twin{})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	client.MaxRetries = 3
	if _, err := client.GetTwin(context.Background(), "twin-1"); err != nil {
		t.Fatalf("retry should succeed: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempts = %d, want 3", attempts)
	}
}

func TestAppendAssertionsBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/assertions/batch" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var payload struct {
			Assertions []StateAssertionInput `json:"assertions"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if len(payload.Assertions) != 2 {
			t.Fatalf("assertions = %d", len(payload.Assertions))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"assertions": []StateAssertion{{ID: "a1"}, {ID: "a2"}}})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := client.AppendAssertions(context.Background(), []StateAssertionInput{{SubjectID: "robot:WH-17", Property: "health"}, {SubjectID: "robot:WH-17", Property: "status"}})
	if err != nil || len(claims) != 2 || claims[1].ID != "a2" {
		t.Fatalf("claims=%#v err=%v", claims, err)
	}
}

func TestGetRelationByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/relations/r1" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(RelationAssertion{ID: "r1", TenantID: "acme"})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	relation, err := client.GetRelation(context.Background(), "r1")
	if err != nil || relation.ID != "r1" {
		t.Fatalf("relation=%#v err=%v", relation, err)
	}
}

func TestLatestClaimAndSnapshotScope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/twins/robot:wh-17/latest/health":
			_ = json.NewEncoder(w).Encode(StateAssertion{StateAssertionInput: StateAssertionInput{Property: "health"}, ID: "a1"})
		case r.URL.Path == "/v1/twins/robot:wh-17/snapshots":
			if r.URL.Query().Get("include_relations") != "false" {
				t.Fatalf("include_relations = %q", r.URL.Query().Get("include_relations"))
			}
			_ = json.NewEncoder(w).Encode(Snapshot{ID: "snap-1", Relations: nil})
		default:
			t.Fatalf("path = %s", r.URL.Path)
		}
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	claim, err := client.LatestClaim(context.Background(), "robot:wh-17", "health")
	if err != nil || claim.Property != "health" {
		t.Fatalf("claim=%#v err=%v", claim, err)
	}
	snapshot, err := client.CreateSnapshotScoped(context.Background(), "robot:wh-17", false, TemporalQuery{})
	if err != nil || snapshot.ID != "snap-1" {
		t.Fatalf("snapshot=%#v err=%v", snapshot, err)
	}
}

func TestGetAssertionByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/assertions/a1" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(StateAssertion{ID: "a1", TenantID: "acme"})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	assertion, err := client.GetAssertion(context.Background(), "a1")
	if err != nil || assertion.ID != "a1" {
		t.Fatalf("assertion=%#v err=%v", assertion, err)
	}
}

func TestSubscriptionDefinitionCRUD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/subscriptions/definitions" && r.Method == http.MethodPost:
			var payload SubscriptionDefinition
			_ = json.NewDecoder(r.Body).Decode(&payload)
			payload.TenantID = "acme"
			_ = json.NewEncoder(w).Encode(payload)
		case r.URL.Path == "/v1/subscriptions/definitions/sub-1" && r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(SubscriptionDefinition{TenantID: "acme", SubscriptionID: "sub-1"})
		case r.URL.Path == "/v1/subscriptions/definitions/sub-1" && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("path=%s method=%s", r.URL.Path, r.Method)
		}
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	created, err := client.CreateSubscriptionDefinition(context.Background(), SubscriptionDefinition{SubscriptionID: "sub-1", Filters: ChangeFilter{Kind: "state_assertion.appended"}})
	if err != nil || created.TenantID != "acme" {
		t.Fatalf("created=%#v err=%v", created, err)
	}
	got, err := client.GetSubscriptionDefinition(context.Background(), "sub-1")
	if err != nil || got.SubscriptionID != "sub-1" {
		t.Fatalf("got=%#v err=%v", got, err)
	}
	if err := client.DeleteSubscriptionDefinition(context.Background(), "sub-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestListAndRevokeSourceBindings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v1/source-bindings" && r.Method == http.MethodGet:
			if r.URL.Query().Get("source") != "sensor:health" {
				t.Fatalf("query = %s", r.URL.RawQuery)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"source_bindings": []SourceBinding{{SourceBindingInput: SourceBindingInput{ID: "b1", Source: "sensor:health"}}}})
		case r.URL.Path == "/v1/source-bindings/b1" && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("path=%s method=%s", r.URL.Path, r.Method)
		}
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	bindings, err := client.ListSourceBindings(context.Background(), "sensor:health")
	if err != nil || len(bindings) != 1 || bindings[0].ID != "b1" {
		t.Fatalf("bindings=%#v err=%v", bindings, err)
	}
	if err := client.RevokeSourceBinding(context.Background(), "b1"); err != nil {
		t.Fatalf("revoke: %v", err)
	}
}

func TestUpdateTwinLifecyclePostsLifecycle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/twins/robot:wh-17/lifecycle" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var payload struct{ Lifecycle string }
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if payload.Lifecycle != "retired" {
			t.Fatalf("lifecycle = %q", payload.Lifecycle)
		}
		_ = json.NewEncoder(w).Encode(Twin{TwinInput: TwinInput{ID: "robot:wh-17", TypeRef: "robot", Lifecycle: "retired"}, TenantID: "acme"})
	}))
	defer server.Close()
	client, err := NewClient(server.URL, "acme")
	if err != nil {
		t.Fatal(err)
	}
	twin, err := client.UpdateTwinLifecycle(context.Background(), "robot:wh-17", "retired")
	if err != nil || twin.Lifecycle != "retired" {
		t.Fatalf("twin=%#v err=%v", twin, err)
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
