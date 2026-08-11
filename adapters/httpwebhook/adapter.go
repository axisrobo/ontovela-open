// Package httpwebhook is a thin ONTOVELA reference adapter that ingests webhook
// HTTP payloads as tenant-scoped, idempotent state assertions through the
// public Go SDK. It never changes state kinds or promotes simulated values.
package httpwebhook

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

var ErrInvalidRequest = errors.New("invalid webhook request")

// IngestRequest is the minimal webhook body accepted by the adapter.
type IngestRequest struct {
	TenantID       string          `json:"tenant_id"`
	IdempotencyKey string          `json:"idempotency_key"`
	SubjectID      string          `json:"subject_id"`
	Property       string          `json:"property"`
	Value          json.RawMessage `json:"value"`
	StateKind      string          `json:"state_kind"`
	EventTime      time.Time       `json:"event_time"`
	Source         string          `json:"source"`
	EvidenceRef    string          `json:"evidence_ref"`
}

func (r IngestRequest) validate() error {
	if r.TenantID == "" || r.IdempotencyKey == "" || r.SubjectID == "" || r.Property == "" || r.Source == "" || r.EvidenceRef == "" || !json.Valid(r.Value) {
		return ErrInvalidRequest
	}
	switch r.StateKind {
	case "observed", "reported", "derived", "inferred", "predicted", "simulated":
	default:
		return ErrInvalidRequest
	}
	return nil
}

// Ingest maps a webhook body to a state assertion and appends it through the
// public Go SDK.
func Ingest(ctx context.Context, client *ontovela.Client, request IngestRequest) (ontovela.StateAssertion, error) {
	if err := request.validate(); err != nil {
		return ontovela.StateAssertion{}, err
	}
	return client.AppendAssertion(ctx, ontovela.StateAssertionInput{
		SubjectID:   request.SubjectID,
		Property:    request.Property,
		Value:       request.Value,
		StateKind:   ontovela.StateKind(request.StateKind),
		EventTime:   request.EventTime,
		Source:      request.Source,
		EvidenceRef: request.EvidenceRef,
	}, request.IdempotencyKey)
}

// Server exposes POST /ingest for direct webhook delivery.
type Server struct {
	Client *ontovela.Client
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.URL.Path != "/ingest" {
		http.NotFound(w, r)
		return
	}
	var request IngestRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if _, err := Ingest(r.Context(), s.Client, request); err != nil {
		writeJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func writeJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
