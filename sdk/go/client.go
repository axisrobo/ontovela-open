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
	MaxRetries int
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

func (c *Client) ListTwinTypes(ctx context.Context) ([]TwinType, error) {
	var result struct {
		TwinTypes []TwinType `json:"twin_types"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/twin-types", nil, nil, "", &result); err != nil {
		return nil, err
	}
	return result.TwinTypes, nil
}

func (c *Client) ReportHeartbeat(ctx context.Context, source string) (SourceHeartbeat, error) {
	var result SourceHeartbeat
	return result, c.doJSON(ctx, http.MethodPost, "/v1/heartbeats", nil, struct {
		Source string `json:"source"`
	}{Source: source}, "", &result)
}

func (c *Client) CreateTwin(ctx context.Context, input TwinInput) (Twin, error) {
	var result Twin
	return result, c.doJSON(ctx, http.MethodPost, "/v1/twins", nil, input, "", &result)
}

func (c *Client) GetTwin(ctx context.Context, twinID string) (Twin, error) {
	var result Twin
	return result, c.doJSON(ctx, http.MethodGet, path.Join("/v1/twins", twinID), nil, nil, "", &result)
}

func (c *Client) UpdateTwinLifecycle(ctx context.Context, twinID, lifecycle string) (Twin, error) {
	var result Twin
	return result, c.doJSON(ctx, http.MethodPost, path.Join("/v1/twins", twinID, "lifecycle"), nil, struct {
		Lifecycle string `json:"lifecycle"`
	}{Lifecycle: lifecycle}, "", &result)
}

func (c *Client) CreateSourceBinding(ctx context.Context, input SourceBindingInput) (SourceBinding, error) {
	var result SourceBinding
	return result, c.doJSON(ctx, http.MethodPost, "/v1/source-bindings", nil, input, "", &result)
}

func (c *Client) ListSourceBindings(ctx context.Context, source string) ([]SourceBinding, error) {
	query := make(url.Values)
	if source != "" {
		query.Set("source", source)
	}
	var result struct {
		SourceBindings []SourceBinding `json:"source_bindings"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/source-bindings", query, nil, "", &result); err != nil {
		return nil, err
	}
	return result.SourceBindings, nil
}

func (c *Client) RevokeSourceBinding(ctx context.Context, bindingID string) error {
	return c.doJSON(ctx, http.MethodDelete, path.Join("/v1/source-bindings", bindingID), nil, nil, "", nil)
}

func (c *Client) AppendAssertion(ctx context.Context, input StateAssertionInput, idempotencyKey string) (StateAssertion, error) {
	var result StateAssertion
	return result, c.doJSON(ctx, http.MethodPost, "/v1/assertions", nil, input, idempotencyKey, &result)
}

func (c *Client) GetRelation(ctx context.Context, relationID string) (RelationAssertion, error) {
	var result RelationAssertion
	return result, c.doJSON(ctx, http.MethodGet, path.Join("/v1/relations", relationID), nil, nil, "", &result)
}

func (c *Client) AppendAssertions(ctx context.Context, assertions []StateAssertionInput) ([]StateAssertion, error) {
	var result struct {
		Assertions []StateAssertion `json:"assertions"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/assertions/batch", nil, struct {
		Assertions []StateAssertionInput `json:"assertions"`
	}{Assertions: assertions}, "", &result); err != nil {
		return nil, err
	}
	return result.Assertions, nil
}

func (c *Client) AppendRelations(ctx context.Context, relations []RelationAssertionInput) ([]RelationAssertion, error) {
	var result struct {
		Relations []RelationAssertion `json:"relations"`
	}
	if err := c.doJSON(ctx, http.MethodPost, "/v1/relations/batch", nil, struct {
		Relations []RelationAssertionInput `json:"relations"`
	}{Relations: relations}, "", &result); err != nil {
		return nil, err
	}
	return result.Relations, nil
}

func (c *Client) GetAssertion(ctx context.Context, assertionID string) (StateAssertion, error) {
	var result StateAssertion
	return result, c.doJSON(ctx, http.MethodGet, path.Join("/v1/assertions", assertionID), nil, nil, "", &result)
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

func (c *Client) ListRelations(ctx context.Context, twinID, predicate, direction string, temporal TemporalQuery) ([]RelationAssertion, error) {
	query := temporal.values()
	if predicate != "" {
		query.Set("predicate", predicate)
	}
	if direction != "" {
		query.Set("direction", direction)
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

func (c *Client) ListSnapshots(ctx context.Context, twinID string, limit int) ([]Snapshot, error) {
	query := make(url.Values)
	query.Set("limit", strconv.Itoa(limit))
	var result struct {
		Snapshots []Snapshot `json:"snapshots"`
	}
	if err := c.doJSON(ctx, http.MethodGet, path.Join("/v1/twins", twinID, "snapshots"), query, nil, "", &result); err != nil {
		return nil, err
	}
	return result.Snapshots, nil
}

func (c *Client) CreateSnapshot(ctx context.Context, twinID string, temporal TemporalQuery) (Snapshot, error) {
	return c.CreateSnapshotScoped(ctx, twinID, true, temporal)
}

func (c *Client) CreateSnapshotScoped(ctx context.Context, twinID string, includeRelations bool, temporal TemporalQuery) (Snapshot, error) {
	query := temporal.values()
	if !includeRelations {
		query.Set("include_relations", "false")
	}
	var result Snapshot
	return result, c.doJSON(ctx, http.MethodPost, path.Join("/v1/twins", twinID, "snapshots"), query, nil, "", &result)
}

func (c *Client) LatestClaim(ctx context.Context, twinID, property string) (StateAssertion, error) {
	var result StateAssertion
	return result, c.doJSON(ctx, http.MethodGet, path.Join("/v1/twins", twinID, "latest", property), nil, nil, "", &result)
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

func (c *Client) ListConflicts(ctx context.Context, status string, limit int) ([]ConflictRecord, error) {
	query := make(url.Values)
	if status != "" {
		query.Set("status", status)
	}
	query.Set("limit", strconv.Itoa(limit))
	var result struct {
		Conflicts []ConflictRecord `json:"conflicts"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/conflicts", query, nil, "", &result); err != nil {
		return nil, err
	}
	return result.Conflicts, nil
}

func (c *Client) CreateSubscriptionDefinition(ctx context.Context, definition SubscriptionDefinition) (SubscriptionDefinition, error) {
	var result SubscriptionDefinition
	return result, c.doJSON(ctx, http.MethodPost, "/v1/subscriptions/definitions", nil, definition, "", &result)
}

func (c *Client) GetSubscriptionDefinition(ctx context.Context, subscriptionID string) (SubscriptionDefinition, error) {
	var result SubscriptionDefinition
	return result, c.doJSON(ctx, http.MethodGet, path.Join("/v1/subscriptions/definitions", subscriptionID), nil, nil, "", &result)
}

func (c *Client) ListSubscriptionDefinitions(ctx context.Context) ([]SubscriptionDefinition, error) {
	var result struct {
		SubscriptionDefinitions []SubscriptionDefinition `json:"subscription_definitions"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/subscriptions/definitions", nil, nil, "", &result); err != nil {
		return nil, err
	}
	return result.SubscriptionDefinitions, nil
}

func (c *Client) DeleteSubscriptionDefinition(ctx context.Context, subscriptionID string) error {
	return c.doJSON(ctx, http.MethodDelete, path.Join("/v1/subscriptions/definitions", subscriptionID), nil, nil, "", nil)
}

func (c *Client) GetSubscriptionOffset(ctx context.Context, consumerID string) (SubscriptionOffset, error) {
	var result SubscriptionOffset
	return result, c.doJSON(ctx, http.MethodGet, path.Join("/v1/subscriptions", consumerID), nil, nil, "", &result)
}

func (c *Client) CommitSubscriptionOffset(ctx context.Context, consumerID string, offset int64) (SubscriptionOffset, error) {
	var result SubscriptionOffset
	return result, c.doJSON(ctx, http.MethodPost, path.Join("/v1/subscriptions", consumerID, "commit"), nil, struct {
		Offset int64 `json:"offset"`
	}{Offset: offset}, "", &result)
}

func (c *Client) SimToReal(ctx context.Context, twinID, property string, temporal TemporalQuery) (SimToRealDelta, error) {
	var result SimToRealDelta
	return result, c.doJSON(ctx, http.MethodGet, path.Join("/v1/twins", twinID, "sim-to-real", property), temporal.values(), nil, "", &result)
}

func (c *Client) ComputeCausalAnalytics(ctx context.Context, twinID string) (CausalAnalytics, error) {
	var result CausalAnalytics
	return result, c.doJSON(ctx, http.MethodGet, path.Join("/v1/twins", twinID, "causal", "analytics"), nil, nil, "", &result)
}

func (c *Client) ComputeCausalLineage(ctx context.Context, twinID string, maxDepth int, temporal TemporalQuery) ([]CausalLink, error) {
	query := temporal.values()
	query.Set("max_depth", strconv.Itoa(maxDepth))
	var result struct {
		CausalLinks []CausalLink `json:"causal_links"`
	}
	if err := c.doJSON(ctx, http.MethodGet, path.Join("/v1/twins", twinID, "causal"), query, nil, "", &result); err != nil {
		return nil, err
	}
	return result.CausalLinks, nil
}

func (c *Client) ComputeImpact(ctx context.Context, twinID string, maxDepth int, predicate string, temporal TemporalQuery) ([]ImpactPath, error) {
	query := temporal.values()
	query.Set("max_depth", strconv.Itoa(maxDepth))
	if predicate != "" {
		query.Set("predicate", predicate)
	}
	var result struct {
		ImpactPaths []ImpactPath `json:"impact_paths"`
	}
	if err := c.doJSON(ctx, http.MethodGet, path.Join("/v1/twins", twinID, "impact"), query, nil, "", &result); err != nil {
		return nil, err
	}
	return result.ImpactPaths, nil
}

func (c *Client) CreateRealityView(ctx context.Context, input RealityViewRequest, temporal TemporalQuery) (RealityView, error) {
	var result RealityView
	return result, c.doJSON(ctx, http.MethodPost, "/v1/reality-views", temporal.values(), input, "", &result)
}

func (c *Client) AuditExportChanges(ctx context.Context, after int64, limit int) ([]ChangeEvent, error) {
	query := url.Values{"after": []string{strconv.FormatInt(after, 10)}, "limit": []string{strconv.Itoa(limit)}}
	var result struct {
		Events []ChangeEvent `json:"events"`
	}
	if err := c.doJSON(ctx, http.MethodGet, "/v1/audit/changes", query, nil, "", &result); err != nil {
		return nil, err
	}
	return result.Events, nil
}

func (c *Client) ListChanges(ctx context.Context, after int64, limit int, filters ChangeFilter) ([]ChangeEvent, error) {
	query := url.Values{"after": []string{strconv.FormatInt(after, 10)}, "limit": []string{strconv.Itoa(limit)}}
	if filters.Kind != "" {
		query.Set("kind", filters.Kind)
	}
	if filters.SubjectID != "" {
		query.Set("subject_id", filters.SubjectID)
	}
	if filters.Property != "" {
		query.Set("property", filters.Property)
	}
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
	var response *http.Response
	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		response, err = client.Do(request)
		if err != nil {
			return err
		}
		if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
			break
		}
		if !c.retryable(response.StatusCode) || attempt == c.MaxRetries {
			defer response.Body.Close()
			var apiError struct {
				Error string `json:"error"`
			}
			_ = json.NewDecoder(response.Body).Decode(&apiError)
			return &APIError{StatusCode: response.StatusCode, Message: apiError.Error}
		}
		_ = response.Body.Close()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(1<<attempt) * 100 * time.Millisecond):
		}
	}
	defer response.Body.Close()
	if output == nil || response.StatusCode == http.StatusNoContent {
		return nil
	}
	return json.NewDecoder(response.Body).Decode(output)
}

func (c *Client) retryable(status int) bool {
	return status == http.StatusTooManyRequests || status >= http.StatusInternalServerError
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
