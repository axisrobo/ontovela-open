// Package coap maps protocol messages to ONTOVELA assertions through the
// shared base payload contract.
package coap

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/axisrobo/ONTOVELA-open/adapters/base"
	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

var ErrInvalidMessage = errors.New("invalid coap message")

// Message is a normalized transport message. Body carries the ONTOVELA payload.
type Message struct {
	Path  string
	Token string
	Body  []byte
}

// Source delivers messages. A coap client library implements it.
type Source interface {
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

// Run consumes messages and appends them as state assertions.
func Run(ctx context.Context, source Source, client *ontovela.Client) error {
	return source.Consume(ctx, func(message Message) error {
		var payload base.Payload
		if err := json.Unmarshal(message.Body, &payload); err != nil {
			return ErrInvalidMessage
		}
		return base.Append(ctx, client, payload)
	})
}
