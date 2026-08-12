// Package base provides the shared payload validation and SDK append path used
// by every protocol adapter, so transport adapters stay small and consistent.
package base

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

var ErrInvalidPayload = errors.New("invalid adapter payload")

// Payload is the normalized message body shared by all protocol adapters.
type Payload struct {
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

// Validate checks the payload without requiring a live core.
func (p Payload) Validate() error {
	if p.TenantID == "" || p.IdempotencyKey == "" || p.SubjectID == "" || p.Property == "" || p.Source == "" || p.EvidenceRef == "" || !json.Valid(p.Value) {
		return fmt.Errorf("%w: missing required fields", ErrInvalidPayload)
	}
	switch p.StateKind {
	case "observed", "reported", "derived", "inferred", "predicted", "simulated":
		return nil
	default:
		return fmt.Errorf("%w: invalid state_kind %q", ErrInvalidPayload, p.StateKind)
	}
}

// Append validates and appends a payload through the core SDK client.
func Append(ctx context.Context, client *ontovela.Client, payload Payload) error {
	if err := payload.Validate(); err != nil {
		return err
	}
	if payload.TenantID != client.TenantID {
		return fmt.Errorf("%w: tenant mismatch", ErrInvalidPayload)
	}
	_, err := client.AppendAssertion(ctx, ontovela.StateAssertionInput{
		SubjectID: payload.SubjectID, Property: payload.Property, Value: payload.Value,
		StateKind: ontovela.StateKind(payload.StateKind), EventTime: payload.EventTime,
		Source: payload.Source, EvidenceRef: payload.EvidenceRef,
	}, payload.IdempotencyKey)
	return err
}
