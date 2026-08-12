# Public API Compatibility

`api/openapi.yaml` is the v0.1 source of truth for ONTOVELA public wire contracts.

- Additive optional fields and endpoints are minor-version changes.
- Removing or changing an existing field, enum value, endpoint, or temporal meaning requires a major version.
- `state_kind`, `event_time`, `system_time`, source, evidence, tenant scope, and idempotency semantics are integrity invariants, not implementation details.
- Enterprise extensions must use explicit namespaces and cannot change the meaning of v0.1 claims, snapshots, digests, or change-feed offsets.
- History reads accept an optional `limit` (1–1000); omitting it returns the full default page.
- `POST /v1/reality-views` with `twin_ids` returns `{views:[...]}` instead of a single view.
- All error responses use the `Error` schema `{error: string}`; non-2xx status semantics are stable.
- Operations endpoints (`/healthz`, `/metrics`, `/v1/version`) are deployment surfaces and are not part of the SDK parity contract.
- Servers may be deployed behind a reverse proxy; base-path handling is an operator configuration, not part of the wire contract.
