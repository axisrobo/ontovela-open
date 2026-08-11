// Package opcua maps OPC UA node reads into ONTOVELA state assertions. An OPC
// UA client library feeds normalized node values in.
package opcua

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

var ErrInvalidNode = errors.New("invalid OPC UA node value")

// NodeValue is a normalized OPC UA node read.
type NodeValue struct {
	TenantID       string          `json:"tenant_id"`
	IdempotencyKey string          `json:"idempotency_key"`
	TwinID         string          `json:"twin_id"`
	Property       string          `json:"property"`
	Value          json.RawMessage `json:"value"`
	Quality        string          `json:"quality"`
	ObservedAt     time.Time       `json:"observed_at"`
	Source         string          `json:"source"`
	EvidenceRef    string          `json:"evidence_ref"`
}

// ToAssertionInput maps a node read to an observed assertion. Poor quality
// reads are rejected so bad data cannot become state.
func ToAssertionInput(node NodeValue) (ontovela.StateAssertionInput, error) {
	if node.TenantID == "" || node.IdempotencyKey == "" || node.TwinID == "" || node.Property == "" || node.Source == "" || node.EvidenceRef == "" || !json.Valid(node.Value) {
		return ontovela.StateAssertionInput{}, fmt.Errorf("%w: missing required fields", ErrInvalidNode)
	}
	if node.Quality != "" && node.Quality != "good" {
		return ontovela.StateAssertionInput{}, fmt.Errorf("%w: quality %q is not good", ErrInvalidNode, node.Quality)
	}
	return ontovela.StateAssertionInput{
		SubjectID: node.TwinID, Property: node.Property, Value: node.Value,
		StateKind: ontovela.Observed, EventTime: node.ObservedAt, Source: node.Source, EvidenceRef: node.EvidenceRef,
	}, nil
}
