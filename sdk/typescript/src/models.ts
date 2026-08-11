export type StateKind = "observed" | "reported" | "derived" | "inferred" | "predicted" | "simulated";

export interface TwinInput {
  id: string;
  type_ref: string;
  lifecycle?: string;
  external_bindings?: unknown;
}

export interface Twin extends TwinInput {
  tenant_id: string;
  created_at?: string;
}

export interface SourceBindingInput {
  id: string;
  source: string;
  property: string;
  authority_rank: number;
  max_lag_seconds: number;
}

export interface SourceBinding extends SourceBindingInput {
  tenant_id: string;
  created_at?: string;
}

export interface StateAssertionInput {
  subject_id: string;
  property: string;
  value: unknown;
  state_kind: StateKind;
  event_time: string;
  source: string;
  confidence?: number;
  evidence_ref: string;
}

export interface StateAssertion extends StateAssertionInput {
  id: string;
  tenant_id: string;
  system_time: string;
}

export interface RelationAssertionInput {
  source_id: string;
  predicate: string;
  target_id: string;
  state_kind: StateKind;
  event_time: string;
  valid_from: string;
  valid_to?: string;
  source: string;
  confidence?: number;
  evidence_ref: string;
}

export interface RelationAssertion extends RelationAssertionInput {
  id: string;
  tenant_id: string;
  system_time: string;
}

export interface ResolvedState {
  tenant_id: string;
  subject_id: string;
  property: string;
  status: "resolved" | "unknown" | "conflicted";
  value?: unknown;
  state_kind?: StateKind;
  source?: string;
  confidence?: number;
  supporting_assertion_id?: string;
  conflicting_assertion_ids?: string[];
  freshness: string;
  resolution_policy: string;
  event_time?: string;
}

export interface RequiredState {
  property: string;
  max_age_seconds: number;
}

export interface RealityViewRequest {
  twin_id: string;
  purpose: string;
  required_state: RequiredState[];
  authorization_ref?: string;
}

export interface RealityViewItem {
  property: string;
  acceptable: boolean;
  state: ResolvedState;
}

export interface RealityView {
  tenant_id: string;
  twin_id: string;
  purpose: string;
  authorization_ref?: string;
  status: "ready" | "stale" | "unknown" | "conflicted";
  items: RealityViewItem[];
}

export interface Snapshot {
  id: string;
  tenant_id: string;
  subject_id: string;
  as_of?: string;
  as_known?: string;
  resolution_policy: string;
  digest: string;
  states: ResolvedState[];
  relations: RelationAssertion[];
  created_at: string;
}

export interface SnapshotDiff {
  from_snapshot_id: string;
  to_snapshot_id: string;
  added_properties: string[];
  removed_properties: string[];
  changed_properties: string[];
  added_relation_ids: string[];
  removed_relation_ids: string[];
}

export interface ChangeEvent {
  offset: number;
  tenant_id: string;
  kind: string;
  subject_id: string;
  property?: string;
  payload: unknown;
  occurred_at: string;
}

export interface SubscriptionOffset {
  tenant_id: string;
  consumer_id: string;
  committed_offset: number;
  updated_at?: string;
}

export interface ConflictRecord {
  tenant_id: string;
  subject_id: string;
  property: string;
  status: "open" | "resolved";
  assertion_ids: string[];
  resolution_policy: string;
  updated_at?: string;
}

export interface SourceHeartbeat {
  tenant_id: string;
  source: string;
  last_heartbeat_at: string;
}

export interface TwinType {
  type_ref: string;
  description: string;
  properties: string[];
  relations: string[];
}

export interface SimToRealDelta {
  tenant_id: string;
  twin_id: string;
  property: string;
  real_state: ResolvedState;
  simulated_state: ResolvedState;
  delta: "match" | "diverges" | "unknown";
}

export interface CausalAnalytics {
  twin_id: string;
  fan_out: number;
  fan_in: number;
  top_targets: Record<string, number>;
}

export interface CausalLink {
  subject_id: string;
  depth: number;
  relation_id: string;
  mechanism?: string;
  target_id: string;
  event_time: string;
  source: string;
  evidence_ref: string;
}

export interface ImpactPath {
  subject_id: string;
  depth: number;
  relation_id: string;
  predicate: string;
  target_id: string;
  state_kind: StateKind;
  source: string;
  evidence_ref: string;
}
