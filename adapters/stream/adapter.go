// Package stream is a broker-neutral ONTOVELA stream-ingest reference adapter.
// Kafka and NATS client libraries can implement Consumer without adding broker
// code to ONTOVELA.
package stream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

var ErrInvalidMessage = errors.New("invalid stream message")

// Message is a raw stream message plus offset metadata.
type Message struct {
	Body      []byte
	Topic     string
	Partition int
	Offset    int64
}

// Consumer delivers stream messages to a handler. Broker clients (Kafka, NATS)
// implement this interface.
type Consumer interface {
	Consume(ctx context.Context, handler func(Message) error) error
}

// MemorySource delivers queued messages for deterministic tests.
type MemorySource struct {
	Messages []Message
}

func (m *MemorySource) Consume(_ context.Context, handler func(Message) error) error {
	for _, message := range m.Messages {
		if err := handler(message); err != nil {
			return err
		}
	}
	return nil
}

// IngestRequest mirrors the webhook adapter's message shape.
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
		return ErrInvalidMessage
	}
	switch r.StateKind {
	case "observed", "reported", "derived", "inferred", "predicted", "simulated":
		return nil
	default:
		return fmt.Errorf("%w: invalid state_kind %q", ErrInvalidMessage, r.StateKind)
	}
}

// ingestMessage maps one stream message to a state assertion and appends it.
func ingestMessage(ctx context.Context, client *ontovela.Client, body []byte) error {
	var request IngestRequest
	if err := json.Unmarshal(body, &request); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidMessage, err)
	}
	if err := request.validate(); err != nil {
		return err
	}
	if request.TenantID != client.TenantID {
		return fmt.Errorf("%w: message tenant %q does not match client tenant %q", ErrInvalidMessage, request.TenantID, client.TenantID)
	}
	_, err := client.AppendAssertion(ctx, ontovela.StateAssertionInput{
		SubjectID:   request.SubjectID,
		Property:    request.Property,
		Value:       request.Value,
		StateKind:   ontovela.StateKind(request.StateKind),
		EventTime:   request.EventTime,
		Source:      request.Source,
		EvidenceRef: request.EvidenceRef,
	}, request.IdempotencyKey)
	return err
}

// Run consumes messages from a source and ingests each into the core. It
// returns the first handler error.
func Run(ctx context.Context, source Consumer, client *ontovela.Client) error {
	return source.Consume(ctx, func(message Message) error {
		return ingestMessage(ctx, client, message.Body)
	})
}
