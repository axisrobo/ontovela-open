// Package effect maps executor EffectRecords into ONTOVELA state assertions.
// It defaults to observed and never emits simulated or predicted state.
package effect

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

var ErrInvalidEffect = errors.New("invalid effect record")

// EffectRecord is a normalized executor effect.
type EffectRecord struct {
	EffectID    string          `json:"effect_id"`
	TenantID    string          `json:"tenant_id"`
	TwinID      string          `json:"twin_id"`
	Property    string          `json:"property"`
	Value       json.RawMessage `json:"value"`
	ObservedAt  time.Time       `json:"observed_at"`
	Source      string          `json:"source"`
	EvidenceRef string          `json:"evidence_ref"`
	StateKind   string          `json:"state_kind,omitempty"`
}

// ToAssertionInput converts an effect into a core assertion input. State kind
// defaults to observed; reported is allowed explicitly.
func ToAssertionInput(effect EffectRecord) (ontovela.StateAssertionInput, error) {
	if effect.EffectID == "" || effect.TenantID == "" || effect.TwinID == "" || effect.Property == "" || effect.Source == "" || effect.EvidenceRef == "" || !json.Valid(effect.Value) {
		return ontovela.StateAssertionInput{}, fmt.Errorf("%w: missing required effect fields", ErrInvalidEffect)
	}
	kind := effect.StateKind
	if kind == "" {
		kind = "observed"
	}
	switch kind {
	case "observed", "reported":
	default:
		return ontovela.StateAssertionInput{}, fmt.Errorf("%w: effect cannot be %q", ErrInvalidEffect, kind)
	}
	return ontovela.StateAssertionInput{
		SubjectID:   effect.TwinID,
		Property:    effect.Property,
		Value:       effect.Value,
		StateKind:   ontovela.StateKind(kind),
		EventTime:   effect.ObservedAt,
		Source:      effect.Source,
		EvidenceRef: effect.EvidenceRef,
	}, nil
}
