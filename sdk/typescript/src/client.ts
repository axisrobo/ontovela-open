import type {
  ChangeEvent,
  ConflictRecord,
  RealityView,
  RealityViewRequest,
  RelationAssertion,
  RelationAssertionInput,
  Snapshot,
  SnapshotDiff,
  SourceBinding,
  SourceBindingInput,
  StateAssertion,
  StateAssertionInput,
  SubscriptionOffset,
  Twin,
  TwinInput,
} from "./models";

export class APIError extends Error {
  readonly serverMessage: string;

  constructor(
    public status: number,
    message: string,
  ) {
    super(`ONTOVELA API ${status}: ${message}`);
    this.name = "APIError";
    this.serverMessage = message;
  }
}

export interface TemporalQuery {
  as_of?: string;
  as_known?: string;
}

interface FetchLike {
  (input: string, init?: RequestInit): Promise<Response>;
}

export interface OntovelaClientOptions {
  baseUrl: string;
  tenantId: string;
  fetch?: FetchLike;
}

export class OntovelaClient {
  private readonly baseUrl: string;
  private readonly tenantId: string;
  private readonly fetchFn: FetchLike;

  constructor(options: OntovelaClientOptions) {
    if (!options.tenantId || !options.tenantId.trim()) {
      throw new Error("tenantId is required");
    }
    this.baseUrl = options.baseUrl.replace(/\/$/, "");
    this.tenantId = options.tenantId;
    this.fetchFn = options.fetch ?? globalThis.fetch.bind(globalThis);
  }

  async reportHeartbeat(source: string): Promise<import("./models").SourceHeartbeat> {
    return this.request<import("./models").SourceHeartbeat>("POST", "/v1/heartbeats", { body: { source } });
  }

  async listTwinTypes(): Promise<import("./models").TwinType[]> {
    const body = await this.request<{ twin_types: import("./models").TwinType[] }>("GET", "/v1/twin-types");
    return body.twin_types;
  }

  async createTwin(input: TwinInput): Promise<Twin> {
    return this.request<Twin>("POST", "/v1/twins", { body: input });
  }

  async getTwin(twinId: string): Promise<Twin> {
    return this.request<Twin>("GET", `/v1/twins/${encodeURIComponent(twinId)}`);
  }

  async createSourceBinding(input: SourceBindingInput): Promise<SourceBinding> {
    return this.request<SourceBinding>("POST", "/v1/source-bindings", { body: input });
  }

  async appendAssertion(input: StateAssertionInput, idempotencyKey: string): Promise<StateAssertion> {
    return this.request<StateAssertion>("POST", "/v1/assertions", { body: input, idempotencyKey });
  }

  async appendRelation(input: RelationAssertionInput, idempotencyKey: string): Promise<RelationAssertion> {
    return this.request<RelationAssertion>("POST", "/v1/relations", { body: input, idempotencyKey });
  }

  async listAssertions(twinId: string, property?: string, temporal?: TemporalQuery): Promise<StateAssertion[]> {
    const query = this.temporalQuery(property === undefined ? {} : { property }, temporal);
    const body = await this.request<{ assertions: StateAssertion[] }>("GET", `/v1/twins/${encodeURIComponent(twinId)}/assertions`, { query });
    return body.assertions;
  }

  async listRelations(twinId: string, predicate?: string, temporal?: TemporalQuery): Promise<RelationAssertion[]> {
    const query = this.temporalQuery(predicate === undefined ? {} : { predicate }, temporal);
    const body = await this.request<{ relations: RelationAssertion[] }>("GET", `/v1/twins/${encodeURIComponent(twinId)}/relations`, { query });
    return body.relations;
  }

  async resolveState(twinId: string, property: string, temporal?: TemporalQuery): Promise<import("./models").ResolvedState> {
    return this.request<import("./models").ResolvedState>(
      "GET",
      `/v1/twins/${encodeURIComponent(twinId)}/state/${encodeURIComponent(property)}`,
      { query: this.temporalQuery({}, temporal) },
    );
  }

  async computeCausalLineage(twinId: string, maxDepth = 5, temporal?: TemporalQuery): Promise<import("./models").CausalLink[]> {
    const query: Record<string, string> = { max_depth: String(maxDepth) };
    Object.assign(query, this.temporalQuery({}, temporal));
    const body = await this.request<{ causal_links: import("./models").CausalLink[] }>("GET", `/v1/twins/${encodeURIComponent(twinId)}/causal`, { query });
    return body.causal_links;
  }

  async computeImpact(twinId: string, options: { maxDepth?: number; predicate?: string; temporal?: TemporalQuery } = {}): Promise<import("./models").ImpactPath[]> {
    const query: Record<string, string> = { max_depth: String(options.maxDepth ?? 5) };
    if (options.predicate !== undefined) {
      query.predicate = options.predicate;
    }
    Object.assign(query, this.temporalQuery({}, options.temporal));
    const body = await this.request<{ impact_paths: import("./models").ImpactPath[] }>("GET", `/v1/twins/${encodeURIComponent(twinId)}/impact`, { query });
    return body.impact_paths;
  }

  async createRealityView(input: RealityViewRequest, temporal?: TemporalQuery): Promise<RealityView> {
    return this.request<RealityView>("POST", "/v1/reality-views", { body: input, query: this.temporalQuery({}, temporal) });
  }

  async listSnapshots(twinId: string, limit = 100): Promise<Snapshot[]> {
    const body = await this.request<{ snapshots: Snapshot[] }>("GET", `/v1/twins/${encodeURIComponent(twinId)}/snapshots`, { query: { limit: String(limit) } });
    return body.snapshots;
  }

  async createSnapshot(twinId: string, temporal?: TemporalQuery): Promise<Snapshot> {
    return this.request<Snapshot>("POST", `/v1/twins/${encodeURIComponent(twinId)}/snapshots`, { query: this.temporalQuery({}, temporal) });
  }

  async getSnapshot(snapshotId: string): Promise<Snapshot> {
    return this.request<Snapshot>("GET", `/v1/snapshots/${encodeURIComponent(snapshotId)}`);
  }

  async verifySnapshot(snapshotId: string): Promise<boolean> {
    const body = await this.request<{ snapshot_id: string; valid: boolean }>("GET", `/v1/snapshots/${encodeURIComponent(snapshotId)}/verify`);
    return body.valid;
  }

  async diffSnapshots(fromSnapshotId: string, toSnapshotId: string): Promise<SnapshotDiff> {
    return this.request<SnapshotDiff>(
      "GET",
      `/v1/snapshots/${encodeURIComponent(fromSnapshotId)}/diff/${encodeURIComponent(toSnapshotId)}`,
    );
  }

  async listChanges(after = 0, limit = 100, filters: { kind?: string; subjectId?: string; property?: string } = {}): Promise<ChangeEvent[]> {
    const query: Record<string, string> = { after: String(after), limit: String(limit) };
    if (filters.kind !== undefined) query.kind = filters.kind;
    if (filters.subjectId !== undefined) query.subject_id = filters.subjectId;
    if (filters.property !== undefined) query.property = filters.property;
    const body = await this.request<{ events: ChangeEvent[] }>("GET", "/v1/changes", { query });
    return body.events;
  }

  async getSubscriptionOffset(consumerId: string): Promise<SubscriptionOffset> {
    return this.request<SubscriptionOffset>("GET", `/v1/subscriptions/${encodeURIComponent(consumerId)}`);
  }

  async commitSubscriptionOffset(consumerId: string, offset: number): Promise<SubscriptionOffset> {
    return this.request<SubscriptionOffset>("POST", `/v1/subscriptions/${encodeURIComponent(consumerId)}/commit`, { body: { offset } });
  }

  async listConflicts(status?: "open" | "resolved", limit = 100): Promise<ConflictRecord[]> {
    const query: Record<string, string> = { limit: String(limit) };
    if (status !== undefined) {
      query.status = status;
    }
    const body = await this.request<{ conflicts: ConflictRecord[] }>("GET", "/v1/conflicts", { query });
    return body.conflicts;
  }

  private temporalQuery(base: Record<string, string>, temporal?: TemporalQuery): Record<string, string> {
    const query = { ...base };
    if (temporal?.as_of) query.as_of = temporal.as_of;
    if (temporal?.as_known) query.as_known = temporal.as_known;
    return query;
  }

  private async request<T>(
    method: string,
    path: string,
    options: { body?: unknown; idempotencyKey?: string; query?: Record<string, string> } = {},
  ): Promise<T> {
    let url = this.baseUrl + path;
    if (options.query && Object.keys(options.query).length > 0) {
      url += "?" + new URLSearchParams(options.query).toString();
    }
    const headers: Record<string, string> = { Accept: "application/json", "X-Tenant-ID": this.tenantId };
    let body: string | undefined;
    if (options.body !== undefined) {
      headers["Content-Type"] = "application/json";
      body = JSON.stringify(options.body);
    }
    if (options.idempotencyKey) {
      headers["Idempotency-Key"] = options.idempotencyKey;
    }
    const response = await this.fetchFn(url, { method, headers, body });
    const raw = await response.text();
    if (!response.ok) {
      let message = "";
      try {
        message = (JSON.parse(raw) as { error?: string }).error ?? "";
      } catch {
        message = "";
      }
      throw new APIError(response.status, message);
    }
    return raw ? (JSON.parse(raw) as T) : (undefined as T);
  }
}
