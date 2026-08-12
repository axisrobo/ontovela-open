// Package edge provides an offline-tolerant spool that buffers state
// assertions locally and replays them in order when connectivity returns.
package edge

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

// Item is a buffered assertion request.
type Item struct {
	TenantID       string          `json:"tenant_id"`
	IdempotencyKey string          `json:"idempotency_key"`
	SubjectID      string          `json:"subject_id"`
	Property       string          `json:"property"`
	Value          json.RawMessage `json:"value"`
	StateKind      string          `json:"state_kind"`
	EventTime      time.Time       `json:"event_time"`
	Source         string          `json:"source"`
	EvidenceRef    string          `json:"evidence_ref"`
	Sequence       int64           `json:"sequence"`
}

// Spool buffers items in append order and replays them safely.
type Spool struct {
	mu      sync.Mutex
	items   []Item
	nextSeq int64
}

func New() *Spool { return &Spool{nextSeq: 1} }

// FileSpool persists items to a JSONL file so they survive process restarts.
type FileSpool struct {
	Spool
	path string
}

// NewFileSpool loads existing items from path and returns a durable spool.
func NewFileSpool(path string) (*FileSpool, error) {
	spool := &FileSpool{path: path}
	spool.nextSeq = 1
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var item Item
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			return nil, err
		}
		spool.items = append(spool.items, item)
		if item.Sequence >= spool.nextSeq {
			spool.nextSeq = item.Sequence + 1
		}
	}
	return spool, nil
}

// Append buffers and appends the item to the JSONL file.
func (f *FileSpool) Append(item Item) int64 {
	sequence := f.Spool.Append(item)
	item.Sequence = sequence
	line, err := json.Marshal(item)
	if err == nil {
		_ = appendLine(f.path, line)
	}
	return sequence
}

func appendLine(path string, line []byte) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(append(line, '\n'))
	return err
}

// Append buffers an item locally without requiring connectivity.
func (s *Spool) Append(item Item) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	item.Sequence = s.nextSeq
	s.items = append(s.items, item)
	s.nextSeq++
	return item.Sequence
}

// Len reports the number of buffered items.
func (s *Spool) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.items)
}

// Flush replays buffered items in order through the SDK. Successfully replayed
// items are removed; a failure stops the flush and retains the remainder.
func (s *Spool) Flush(ctx context.Context, client *ontovela.Client) error {
	for {
		s.mu.Lock()
		if len(s.items) == 0 {
			s.mu.Unlock()
			return nil
		}
		item := s.items[0]
		s.mu.Unlock()

		_, err := client.AppendAssertion(ctx, ontovela.StateAssertionInput{
			SubjectID:   item.SubjectID,
			Property:    item.Property,
			Value:       item.Value,
			StateKind:   ontovela.StateKind(item.StateKind),
			EventTime:   item.EventTime,
			Source:      item.Source,
			EvidenceRef: item.EvidenceRef,
		}, item.IdempotencyKey)
		if err != nil {
			return err
		}

		s.mu.Lock()
		s.items = s.items[1:]
		s.mu.Unlock()
	}
}
