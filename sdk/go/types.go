// Package ontovela is the public Go client for the ONTOVELA v0.1 API.
package ontovela

import (
	"encoding/json"
	"time"
)

type StateKind string

const (
	Observed  StateKind = "observed"
	Reported  StateKind = "reported"
	Derived   StateKind = "derived"
	Inferred  StateKind = "inferred"
	Predicted StateKind = "predicted"
	Simulated StateKind = "simulated"
)

type TwinInput struct {
	ID        string          `json:"id"`
	TypeRef   string          `json:"type_ref"`
	Lifecycle string          `json:"lifecycle,omitempty"`
	Bindings  json.RawMessage `json:"external_bindings,omitempty"`
}

type Twin struct {
	TwinInput
	TenantID  string    `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
}

type SourceBindingInput struct {
	ID            string `json:"id"`
	Source        string `json:"source"`
	Property      string `json:"property"`
	AuthorityRank int    `json:"authority_rank"`
	MaxLagSeconds int    `json:"max_lag_seconds"`
}

type SourceBinding struct {
	SourceBindingInput
	TenantID  string    `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
}

type StateAssertionInput struct {
	SubjectID   string          `json:"subject_id"`
	Property    string          `json:"property"`
	Value       json.RawMessage `json:"value"`
	StateKind   StateKind       `json:"state_kind"`
	EventTime   time.Time       `json:"event_time"`
	Source      string          `json:"source"`
	Confidence  *float64        `json:"confidence,omitempty"`
	EvidenceRef string          `json:"evidence_ref"`
}

type StateAssertion struct {
	StateAssertionInput
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	SystemTime time.Time `json:"system_time"`
}

type RelationAssertionInput struct {
	SourceID    string     `json:"source_id"`
	Predicate   string     `json:"predicate"`
	TargetID    string     `json:"target_id"`
	StateKind   StateKind  `json:"state_kind"`
	EventTime   time.Time  `json:"event_time"`
	ValidFrom   time.Time  `json:"valid_from"`
	ValidTo     *time.Time `json:"valid_to,omitempty"`
	Source      string     `json:"source"`
	Confidence  *float64   `json:"confidence,omitempty"`
	EvidenceRef string     `json:"evidence_ref"`
}

type RelationAssertion struct {
	RelationAssertionInput
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	SystemTime time.Time `json:"system_time"`
}

type TemporalQuery struct {
	AsOf    *time.Time
	AsKnown *time.Time
}

type ResolvedState struct {
	TenantID                string          `json:"tenant_id"`
	SubjectID               string          `json:"subject_id"`
	Property                string          `json:"property"`
	Status                  string          `json:"status"`
	Value                   json.RawMessage `json:"value,omitempty"`
	StateKind               StateKind       `json:"state_kind,omitempty"`
	Source                  string          `json:"source,omitempty"`
	Confidence              *float64        `json:"confidence,omitempty"`
	SupportingAssertionID   string          `json:"supporting_assertion_id,omitempty"`
	ConflictingAssertionIDs []string        `json:"conflicting_assertion_ids,omitempty"`
	Freshness               string          `json:"freshness"`
	ResolutionPolicy        string          `json:"resolution_policy"`
	EventTime               time.Time       `json:"event_time,omitempty"`
}

type RequiredState struct {
	Property      string `json:"property"`
	MaxAgeSeconds int    `json:"max_age_seconds"`
}

type RealityViewRequest struct {
	TwinID        string          `json:"twin_id"`
	Purpose       string          `json:"purpose"`
	RequiredState []RequiredState `json:"required_state"`
}

type RealityViewItem struct {
	Property   string        `json:"property"`
	Acceptable bool          `json:"acceptable"`
	State      ResolvedState `json:"state"`
}

type RealityView struct {
	TenantID string            `json:"tenant_id"`
	TwinID   string            `json:"twin_id"`
	Purpose  string            `json:"purpose"`
	Status   string            `json:"status"`
	Items    []RealityViewItem `json:"items"`
}

type Snapshot struct {
	ID               string              `json:"id"`
	TenantID         string              `json:"tenant_id"`
	SubjectID        string              `json:"subject_id"`
	AsOf             *time.Time          `json:"as_of,omitempty"`
	AsKnown          *time.Time          `json:"as_known,omitempty"`
	ResolutionPolicy string              `json:"resolution_policy"`
	Digest           string              `json:"digest"`
	States           []ResolvedState     `json:"states"`
	Relations        []RelationAssertion `json:"relations"`
	CreatedAt        time.Time           `json:"created_at"`
}

type SnapshotDiff struct {
	FromSnapshotID     string   `json:"from_snapshot_id"`
	ToSnapshotID       string   `json:"to_snapshot_id"`
	AddedProperties    []string `json:"added_properties"`
	RemovedProperties  []string `json:"removed_properties"`
	ChangedProperties  []string `json:"changed_properties"`
	AddedRelationIDs   []string `json:"added_relation_ids"`
	RemovedRelationIDs []string `json:"removed_relation_ids"`
}

type ChangeEvent struct {
	Offset     int64           `json:"offset"`
	TenantID   string          `json:"tenant_id"`
	Kind       string          `json:"kind"`
	SubjectID  string          `json:"subject_id"`
	Property   string          `json:"property,omitempty"`
	Payload    json.RawMessage `json:"payload"`
	OccurredAt time.Time       `json:"occurred_at"`
}

type SubscriptionOffset struct {
	TenantID        string    `json:"tenant_id"`
	ConsumerID      string    `json:"consumer_id"`
	CommittedOffset int64     `json:"committed_offset"`
	UpdatedAt       time.Time `json:"updated_at,omitempty"`
}

type ConflictRecord struct {
	TenantID         string    `json:"tenant_id"`
	SubjectID        string    `json:"subject_id"`
	Property         string    `json:"property"`
	Status           string    `json:"status"`
	AssertionIDs     []string  `json:"assertion_ids"`
	ResolutionPolicy string    `json:"resolution_policy"`
	UpdatedAt        time.Time `json:"updated_at,omitempty"`
}
