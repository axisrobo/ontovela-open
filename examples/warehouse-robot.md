# Warehouse Robot v0.1 Example

All requests include `X-Tenant-ID`. Operational writes also require an `Idempotency-Key` and a prior source binding.

```bash
curl -X POST http://localhost:8080/v1/twins \
  -H 'X-Tenant-ID: acme' -H 'Content-Type: application/json' \
  -d '{"id":"robot:WH-17","type_ref":"robot"}'

curl -X POST http://localhost:8080/v1/source-bindings \
  -H 'X-Tenant-ID: acme' -H 'Content-Type: application/json' \
  -d '{"id":"uwb-location","source":"sensor:uwb-gateway-3","property":"location","authority_rank":100,"max_lag_seconds":5}'

curl -X POST http://localhost:8080/v1/assertions \
  -H 'X-Tenant-ID: acme' -H 'Idempotency-Key: uwb-9f2' -H 'Content-Type: application/json' \
  -d '{"subject_id":"robot:WH-17","property":"location","value":[12.44,8.03,0.0],"state_kind":"observed","event_time":"2026-08-11T10:00:00Z","source":"sensor:uwb-gateway-3","confidence":0.982,"evidence_ref":"harmovela:event/9f2"}'

curl -H 'X-Tenant-ID: acme' \
  http://localhost:8080/v1/twins/robot:WH-17/state/location
```

`simulated` assertions can be retained in the ledger for a PEIRAVELA scope, but the resolved-state and Reality Snapshot paths exclude them.
