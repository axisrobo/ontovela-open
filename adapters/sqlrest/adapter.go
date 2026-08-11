// Package sqlrest is a pull-based ONTOVELA reference adapter that polls a SQL
// or REST source and ingests changed rows as state assertions.
package sqlrest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

var ErrInvalidRow = errors.New("invalid source row")

// Cursor is an opaque position in the source.
type Cursor string

// Row is a normalized source record.
type Row struct {
	Cursor         Cursor          `json:"cursor"`
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

// RowSource fetches changed rows after a cursor. SQL queries and REST
// endpoints implement this interface.
type RowSource interface {
	Fetch(ctx context.Context, after Cursor) ([]Row, Cursor, error)
}

// Poller ingests rows and advances the cursor after success.
type Poller struct {
	Source RowSource
	Client *ontovela.Client
}

// PollOnce runs a single fetch and ingestion pass.
func (p *Poller) PollOnce(ctx context.Context, after Cursor) (Cursor, error) {
	rows, next, err := p.Source.Fetch(ctx, after)
	if err != nil {
		return after, err
	}
	for _, row := range rows {
		if err := ingestRow(ctx, p.Client, row); err != nil {
			return after, err
		}
	}
	return next, nil
}

func ingestRow(ctx context.Context, client *ontovela.Client, row Row) error {
	if row.TenantID != client.TenantID {
		return fmt.Errorf("%w: row tenant %q does not match client tenant %q", ErrInvalidRow, row.TenantID, client.TenantID)
	}
	switch row.StateKind {
	case "observed", "reported", "derived", "inferred", "predicted", "simulated":
	default:
		return fmt.Errorf("%w: invalid state_kind %q", ErrInvalidRow, row.StateKind)
	}
	_, err := client.AppendAssertion(ctx, ontovela.StateAssertionInput{
		SubjectID:   row.SubjectID,
		Property:    row.Property,
		Value:       row.Value,
		StateKind:   ontovela.StateKind(row.StateKind),
		EventTime:   row.EventTime,
		Source:      row.Source,
		EvidenceRef: row.EvidenceRef,
	}, row.IdempotencyKey)
	return err
}
