// Package live verifies the five stable source classes against a real core
// backed by PostgreSQL, driving each adapter's real mapping path. Run
// scripts/run-live-core.ps1 first, then:
//   $env:ONTOVELA_LIVE='1'; $env:GOWORK='off'; go test ./... -v
package live

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	effect "github.com/axisrobo/ONTOVELA-open/adapters/effect"
	harmovela "github.com/axisrobo/ONTOVELA-open/adapters/harmovela"
	httpwebhook "github.com/axisrobo/ONTOVELA-open/adapters/httpwebhook"
	mqtt "github.com/axisrobo/ONTOVELA-open/adapters/mqtt"
	opcua "github.com/axisrobo/ONTOVELA-open/adapters/opcua"
	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

const tenantID = "acme"

// liveClient returns an SDK client pinned to the live core, or skips when the
// live suite is not requested or the core is unreachable.
func liveClient(t *testing.T) *ontovela.Client {
	t.Helper()
	if os.Getenv("ONTOVELA_LIVE") != "1" {
		t.Skip("ONTOVELA_LIVE != 1; skipping live source-class verification")
	}
	baseURL := os.Getenv("ONTOVELA_LIVE_CORE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8080"
	}
	client, err := ontovela.NewClient(baseURL, tenantID)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	probe, err := client.ListTwinTypes(ctx)
	if err != nil || len(probe) == 0 {
		t.Skipf("core unreachable at %s (start scripts/run-live-core.ps1); probe=%v err=%v", baseURL, probe, err)
	}
	return client
}

// seedTwin creates a twin plus a source binding for the property.
func seedTwin(t *testing.T, client *ontovela.Client, twinID, property, source string, authorityRank int) {
	t.Helper()
	ctx := context.Background()
	if _, err := client.CreateTwin(ctx, ontovela.TwinInput{ID: twinID, TypeRef: "robot"}); err != nil {
		t.Fatalf("create twin: %v", err)
	}
	if _, err := client.CreateSourceBinding(ctx, ontovela.SourceBindingInput{
		ID: "b-" + twinID, Source: source, Property: property, AuthorityRank: authorityRank, MaxLagSeconds: 600,
	}); err != nil {
		t.Fatalf("create binding: %v", err)
	}
}

// assertResolvedKind resolves a property and verifies the resolved state kind.
func assertResolvedKind(t *testing.T, client *ontovela.Client, twinID, property string, want ontovela.StateKind) {
	t.Helper()
	ctx := context.Background()
	resolved, err := client.ResolveState(ctx, twinID, property, ontovela.TemporalQuery{})
	if err != nil {
		t.Fatalf("resolve %s/%s: %v", twinID, property, err)
	}
	if resolved.Status != "resolved" {
		t.Fatalf("resolve %s/%s status=%s want resolved", twinID, property, resolved.Status)
	}
	if resolved.StateKind != want {
		t.Fatalf("resolve %s/%s kind=%s want %s (no promotion)", twinID, property, resolved.StateKind, want)
	}
}

func TestLiveEffectSourceClass(t *testing.T) {
	client := liveClient(t)
	seedTwin(t, client, "robot:effect-1", "health", "kinetovela:robot", 30)
	record := effect.EffectRecord{
		EffectID: "eff-1", TenantID: tenantID, TwinID: "robot:effect-1",
		Property: "health", Value: json.RawMessage(`"ready"`), ObservedAt: time.Now().UTC(),
		Source: "kinetovela:robot", EvidenceRef: "effect/eff-1",
	}
	// Real adapter mapping: EffectRecord -> StateAssertionInput.
	input, err := effect.ToAssertionInput(record)
	if err != nil {
		t.Fatalf("effect.ToAssertionInput: %v", err)
	}
	if _, err := client.AppendAssertion(context.Background(), input, "k-eff-1"); err != nil {
		t.Fatalf("append effect: %v", err)
	}
	assertResolvedKind(t, client, "robot:effect-1", "health", ontovela.Observed)
}

func TestLiveMqttSourceClass(t *testing.T) {
	client := liveClient(t)
	seedTwin(t, client, "robot:mqtt-1", "health", "mqtt:broker", 20)
	payload := fmt.Sprintf(`{"tenant_id":%q,"idempotency_key":"k-mqtt-1","subject_id":"robot:mqtt-1","property":"health","value":"online","state_kind":"observed","event_time":%q,"source":"mqtt:broker","evidence_ref":"mqtt/1"}`, tenantID, time.Now().UTC().Format(time.RFC3339))
	// Real adapter path: mqtt.Run consumes one broker-shaped message and appends.
	source := &mqtt.MemorySource{Messages: []mqtt.Message{{Topic: "ontovela/state/robot/mqtt-1", Payload: []byte(payload), QOS: 1}}}
	if err := mqtt.Run(context.Background(), source, []string{"ontovela/state/+/+"}, client); err != nil {
		t.Fatalf("mqtt.Run: %v", err)
	}
	assertResolvedKind(t, client, "robot:mqtt-1", "health", ontovela.Observed)
}

func TestLiveHttpWebhookSourceClass(t *testing.T) {
	client := liveClient(t)
	seedTwin(t, client, "robot:webhook-1", "health", "erp:webhook", 10)
	// Real adapter path: the webhook Server's /ingest handler appends via the SDK.
	adapter := &httpwebhook.Server{Client: client}
	handler := httptest.NewServer(adapter)
	defer handler.Close()
	body := fmt.Sprintf(`{"tenant_id":%q,"idempotency_key":"k-webhook-1","subject_id":"robot:webhook-1","property":"health","value":"ok","state_kind":"observed","event_time":%q,"source":"erp:webhook","evidence_ref":"webhook/1"}`, tenantID, time.Now().UTC().Format(time.RFC3339))
	response, err := http.Post(handler.URL+"/ingest", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("webhook post: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusAccepted {
		t.Fatalf("webhook ingest status %d", response.StatusCode)
	}
	assertResolvedKind(t, client, "robot:webhook-1", "health", ontovela.Observed)
}

func TestLiveOpcuaSourceClass(t *testing.T) {
	client := liveClient(t)
	seedTwin(t, client, "robot:opcua-1", "health", "opcua:plc", 5)
	node := opcua.NodeValue{
		TenantID: tenantID, IdempotencyKey: "k-opcua-1", TwinID: "robot:opcua-1",
		Property: "health", Value: json.RawMessage(`"good"`), Quality: "good",
		ObservedAt: time.Now().UTC(), Source: "opcua:plc", EvidenceRef: "opcua/1",
	}
	// Real adapter mapping: NodeValue -> StateAssertionInput (quality-gated).
	input, err := opcua.ToAssertionInput(node)
	if err != nil {
		t.Fatalf("opcua.ToAssertionInput: %v", err)
	}
	if _, err := client.AppendAssertion(context.Background(), input, "k-opcua-1"); err != nil {
		t.Fatalf("append opcua: %v", err)
	}
	assertResolvedKind(t, client, "robot:opcua-1", "health", ontovela.Observed)
}

func TestLiveHarmovelaSourceClass(t *testing.T) {
	client := liveClient(t)
	seedTwin(t, client, "robot:ev-1", "health", "harmovela:evidence", 15)
	envelope := fmt.Sprintf(`{"spec_version":"0.2","id":"ev-1","type":"sensor.reading","source":"harmovela:evidence","created_at":%q,"payload":{"health":"ready"}}`, time.Now().UTC().Format(time.RFC3339))
	evidence := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(envelope))
	}))
	defer evidence.Close()
	// Real adapter path: harmovela.Client validates the envelope over HTTP.
	evidenceClient, err := harmovela.NewClient(evidence.URL, nil)
	if err != nil {
		t.Fatalf("harmovela.NewClient: %v", err)
	}
	if _, err := evidenceClient.FetchAndValidate(context.Background(), "harmovela:event/ev-1"); err != nil {
		t.Fatalf("FetchAndValidate: %v", err)
	}
	// Evidence validated: append the reading as an observed assertion.
	if _, err := client.AppendAssertion(context.Background(), ontovela.StateAssertionInput{
		SubjectID: "robot:ev-1", Property: "health", Value: json.RawMessage(`"ready"`),
		StateKind: ontovela.Observed, EventTime: time.Now().UTC(),
		Source: "harmovela:evidence", EvidenceRef: "harmovela:event/ev-1",
	}, "k-ev-1"); err != nil {
		t.Fatalf("append harmovela: %v", err)
	}
	assertResolvedKind(t, client, "robot:ev-1", "health", ontovela.Observed)
}
