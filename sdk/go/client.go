package ontovela

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

// Client is safe for concurrent use when its HTTPClient is safe for concurrent use.
type Client struct {
	BaseURL    *url.URL
	TenantID   string
	HTTPClient *http.Client
}

func NewClient(baseURL, tenantID string) (*Client, error) {
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid base URL %q", baseURL)
	}
	if strings.TrimSpace(tenantID) == "" {
		return nil, fmt.Errorf("tenant ID is required")
	}
	return &Client{BaseURL: parsed, TenantID: tenantID, HTTPClient: http.DefaultClient}, nil
}

func (c *Client) CreateTwin(ctx context.Context, input TwinInput) (Twin, error) {
	var result Twin
	return result, c.doJSON(ctx, http.MethodPost, "/v1/twins", nil, input, "", &result)
}

func (c *Client) GetTwin(ctx context.Context, twinID string) (Twin, error) {
	var result Twin
	return result, c.doJSON(ctx, http.MethodGet, path.Join("/v1/twins", twinID), nil, nil, "", &result)
}

func (c *Client) CreateSourceBinding(ctx context.Context, input SourceBindingInput) (SourceBinding, error) {
	var result SourceBinding
	return result, c.doJSON(ctx, http.MethodPost, "/v1/source-bindings", nil, input, "", &result)
}

func (c *Client) AppendAssertion(ctx context.Context, input StateAssertionInput, idempotencyKey string) (StateAssertion, error) {
	var result StateAssertion
	return result, c.doJSON(ctx, http.MethodPost, "/v1/assertions", nil, input, idempotencyKey, &result)
}

func (c *Client) AppendRelation(ctx context.Context, input RelationAssertionInput, idempotencyKey string) (RelationAssertion, error) {
	var result RelationAssertion
	return result, c.doJSON(ctx, http.MethodPost, "/v1/relations", nil, input, idempotencyKey, &result)
}

func (c *Client) ListAssertions(ctx context.Context, twinID, property string, temporal TemporalQuery) ([]StateAssertion, error) {
	query := temporal.values()
	if property != "" {
		query.Set("property", property)
	}
	var result struct {
		Assertions []StateAssertion `json:"assertions"`
	}
	if err := c.doJSON(ctx, http.MethodGet, path.Join("/v1/twins", twinID, "assertions"), query, nil, "", &result); err != nil {
		return nil, err
	}
	return result.Assertions, nil
}

func (c *Client) ListRelations(ctx context.Context, twinID, predicate string, temporal TemporalQuery) ([]RelationAssertion, error) {
	query := temporal.values()
	if predicate != "" {
		query.Set("predicate", predicate)
	}
	var result struct {
		Relations []RelationAssertion `json:"relations"`
	}
	if err := c.doJSON(ctx, http.MethodGet, path.Join("/v1/twins", twinID, "relations"), query, nil, "", &result); err != nil {
		return nil, err
	}
	return result.Relations, nil
}

func (c *Client) ResolveState(ctx context.Context, twinID, property string, temporal TemporalQuery) (ResolvedState, error) {
	var result ResolvedState
	return result, c.doJSON(ctx, http.MethodGet, path.Join("/v1/twins", twinID, "state", property), temporal.values(), nil, "", &result)
}

func (c *Client) CreateSnapshot(ctx context.Context, twinID string, temporal TemporalQuery) (Snapshot, error) {
	var result Snapshot
	return result, c.doJSON(ctx, http.MethodPost, path.Join("/v1/twins", twinID, "snapshots"), temporal.values(), nil, "", &result)
}

func (c *Client) GetSnapshot(ctx context.Context, snapshotID string) (Snapshot, error) {
	var result Snapshot
	return result, c.doJSON(ctx, http.MethodGet, path.Join("/v1/snapshots", snapshotID), nil, nil, "", &result)
}

func (c *Client) VerifySnapshot(ctx context.Context, snapshotID string) (bool, error) {
	var result struct {
		SnapshotID string `json:"snapshot_id"`
		Valid      bool   `json:"valid"`
	}
	if err := c.doJSON(ctx, http.MethodGet, path.Join("/v1/snapshots", snapshotID, "verify"), nil, nil, "", &result); err != nil {
		return false, err
	}
	return result.Valid, nil
}

func (c *Client) DiffSnapshots(ctx context.Context, fromSnapshotID, toSnapshotID string) (SnapshotDiff, error) {
	var result SnapshotDiff
	return result, c.doJSON(ctx, http.MethodGet, path.Join("/v1/snapshots", fromSnapshotID, "diff", toSnapshotID), nil, nil, "", &result)
}

func (c *Client) CreateRealityView(ctx context.Context, input RealityViewRequest, temporal TemporalQuery) (RealityView, error) {
	var result RealityView
	return result, c.doJSON(ctx, http.MethodPost, "/v1/reality-views", temporal.values(), input, "", &result)
}

func (c *Client) ListChanges(ctx context.Context, after int64, limit int) ([]ChangeEvent, error) {
	query := url.Values{"after": []string{strconv.FormatInt(after, 10)}, "limit": []string{strconv.Itoa(limit)}}
	var result struct {
		Events []ChangeEvent `json:"events"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/changes", query, nil, "", &result); err != nil {
		return nil, err
	}
	return result.Events, nil
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string { return fmt.Sprintf("ONTOVELA API %d: %s", e.StatusCode, e.Message) }

func (c *Client) doJSON(ctx context.Context, method, endpoint string, query url.Values, input any, idempotencyKey string, output any) error {
	if c == nil || c.BaseURL == nil {
		return fmt.Errorf("nil client")
	}
	relative, err := url.Parse(endpoint)
	if err != nil {
		return err
	}
	target := c.BaseURL.ResolveReference(relative)
	target.RawQuery = query.Encode()
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, target.String(), body)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("X-Tenant-ID", c.TenantID)
	if input != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if idempotencyKey != "" {
		request.Header.Set("Idempotency-Key", idempotencyKey)
	}
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		var apiError struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(response.Body).Decode(&apiError)
		return &APIError{StatusCode: response.StatusCode, Message: apiError.Error}
	}
	if output == nil || response.StatusCode == http.StatusNoContent {
		return nil
	}
	return json.NewDecoder(response.Body).Decode(output)
}

func (q TemporalQuery) values() url.Values {
	values := make(url.Values)
	if q.AsOf != nil {
		values.Set("as_of", q.AsOf.Format(time.RFC3339Nano))
	}
	if q.AsKnown != nil {
		values.Set("as_known", q.AsKnown.Format(time.RFC3339Nano))
	}
	return values
}
