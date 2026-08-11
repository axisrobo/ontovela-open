// Package mqtt is a broker-neutral MQTT-style topic ingestion adapter. An MQTT
// client library implements TopicSource; the adapter maps messages to claims.
package mqtt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

var ErrInvalidMessage = errors.New("invalid MQTT message")

// Message is an MQTT-style topic message.
type Message struct {
	Topic    string
	Payload  []byte
	QOS      int
	Retained bool
}

// TopicSource delivers messages for topics. MQTT client libraries implement it.
type TopicSource interface {
	Consume(ctx context.Context, topics []string, handler func(Message) error) error
}

// MemorySource delivers queued messages for deterministic tests.
type MemorySource struct {
	Messages []Message
}

func (m *MemorySource) Consume(_ context.Context, _ []string, handler func(Message) error) error {
	for _, message := range m.Messages {
		if err := handler(message); err != nil {
			return err
		}
	}
	return nil
}

// Payload is the JSON body of an MQTT message.
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

func validate(payload Payload) error {
	if payload.TenantID == "" || payload.IdempotencyKey == "" || payload.SubjectID == "" || payload.Property == "" || payload.Source == "" || payload.EvidenceRef == "" || !json.Valid(payload.Value) {
		return ErrInvalidMessage
	}
	switch payload.StateKind {
	case "observed", "reported", "derived", "inferred", "predicted", "simulated":
		return nil
	default:
		return fmt.Errorf("%w: invalid state_kind %q", ErrInvalidMessage, payload.StateKind)
	}
}

// Run consumes MQTT messages whose topic carries a robot state, maps them to
// assertions, and appends through the SDK.
func Run(ctx context.Context, source TopicSource, topics []string, client *ontovela.Client) error {
	return source.Consume(ctx, topics, func(message Message) error {
		var payload Payload
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidMessage, err)
		}
		if err := validate(payload); err != nil {
			return err
		}
		if payload.TenantID != client.TenantID {
			return fmt.Errorf("%w: tenant mismatch", ErrInvalidMessage)
		}
		_, err := client.AppendAssertion(ctx, ontovela.StateAssertionInput{
			SubjectID:   payload.SubjectID,
			Property:    payload.Property,
			Value:       payload.Value,
			StateKind:   ontovela.StateKind(payload.StateKind),
			EventTime:   payload.EventTime,
			Source:      payload.Source,
			EvidenceRef: payload.EvidenceRef,
		}, payload.IdempotencyKey)
		return err
	})
}

// TopicFromID is a helper that derives a state topic from a twin ID.
func TopicFromID(twinID string) string {
	return "ontovela/state/" + strings.ReplaceAll(twinID, ":", "/")
}
