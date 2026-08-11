# Public API Compatibility

`api/openapi.yaml` is the v0.1 source of truth for ONTOVELA public wire contracts.

- Additive optional fields and endpoints are minor-version changes.
- Removing or changing an existing field, enum value, endpoint, or temporal meaning requires a major version.
- `state_kind`, `event_time`, `system_time`, source, evidence, tenant scope, and idempotency semantics are integrity invariants, not implementation details.
- Enterprise extensions must use explicit namespaces and cannot change the meaning of v0.1 claims, snapshots, digests, or change-feed offsets.
