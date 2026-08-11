// Package harmovela is a self-contained ONTOVELA reference adapter that
// resolves and validates `harmovela:event/<id>` evidence references against
// the public Harmovela event contract.
package harmovela

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const evidenceScheme = "harmovela:event/"

var (
	ErrInvalidEvidenceRef = errors.New("invalid evidence reference")
	ErrEvidenceNotFound   = errors.New("evidence event not found")
	ErrInvalidEnvelope    = errors.New("invalid Harmovela event envelope")
)

// EvidenceRecord is the normalized evidence result of a validated event.
type EvidenceRecord struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Source    string          `json:"source"`
	CreatedAt time.Time       `json:"created_at"`
	Envelope  json.RawMessage `json:"envelope"`
}

// Client resolves and validates Harmovela event evidence.
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// NewClient returns a Client with the supplied base URL and an HTTP client.
func NewClient(baseURL string, httpClient *http.Client) (*Client, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid Harmovela base URL %q", baseURL)
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{BaseURL: strings.TrimRight(baseURL, "/"), HTTP: httpClient}, nil
}

// ParseEvidenceRef accepts `harmovela:event/<id>` and returns the event ID.
func ParseEvidenceRef(ref string) (string, error) {
	if !strings.HasPrefix(ref, evidenceScheme) {
		return "", fmt.Errorf("%w: expected %s<id>", ErrInvalidEvidenceRef, evidenceScheme)
	}
	id := strings.TrimPrefix(ref, evidenceScheme)
	if id == "" || strings.ContainsAny(id, " /?#") {
		return "", fmt.Errorf("%w: empty or unsafe event id", ErrInvalidEvidenceRef)
	}
	return id, nil
}

// FetchAndValidate resolves an evidence reference and validates its envelope.
func (c *Client) FetchAndValidate(ctx context.Context, ref string) (EvidenceRecord, error) {
	id, err := ParseEvidenceRef(ref)
	if err != nil {
		return EvidenceRecord{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/events/"+url.PathEscape(id), nil)
	if err != nil {
		return EvidenceRecord{}, err
	}
	request.Header.Set("Accept", "application/json")
	response, err := c.HTTP.Do(request)
	if err != nil {
		return EvidenceRecord{}, err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound {
		return EvidenceRecord{}, fmt.Errorf("%w: %s", ErrEvidenceNotFound, id)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return EvidenceRecord{}, fmt.Errorf("evidence fetch returned %d", response.StatusCode)
	}
	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return EvidenceRecord{}, err
	}
	var envelope map[string]any
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return EvidenceRecord{}, fmt.Errorf("%w: %v", ErrInvalidEnvelope, err)
	}
	if validationErrors := validateEnvelope(envelope); len(validationErrors) > 0 {
		return EvidenceRecord{}, fmt.Errorf("%w: %s", ErrInvalidEnvelope, strings.Join(validationErrors, "; "))
	}
	created, _ := time.Parse(time.RFC3339, envelope["created_at"].(string))
	return EvidenceRecord{
		ID:        envelope["id"].(string),
		Type:      envelope["type"].(string),
		Source:    envelope["source"].(string),
		CreatedAt: created,
		Envelope:  json.RawMessage(payload),
	}, nil
}

// validateEnvelope mirrors the public Harmovela event contract fields without
// depending on implementation internals.
func validateEnvelope(value map[string]any) []string {
	var validationErrors []string
	if value == nil {
		return []string{"event must be a JSON object"}
	}
	if version, ok := value["spec_version"].(string); !ok || version != "0.2" {
		validationErrors = append(validationErrors, "spec_version must be 0.2")
	}
	for _, field := range []string{"id", "type", "source", "created_at"} {
		if s, ok := value[field].(string); !ok || s == "" {
			validationErrors = append(validationErrors, field+" must be a non-empty string")
		}
	}
	if _, ok := value["payload"]; !ok {
		validationErrors = append(validationErrors, "payload is required")
	}
	if created, ok := value["created_at"].(string); ok {
		if _, err := time.Parse(time.RFC3339, created); err != nil {
			validationErrors = append(validationErrors, "created_at must be RFC3339")
		}
	}
	return validationErrors
}
