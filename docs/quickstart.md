# ONTOVELA Quickstart

Get a local ONTOVELA instance running in under five minutes.

## Option A: Local developer binary

```powershell
# From the ONTOVELA core repository
cd backend
GOWORK=off go build -o bin/ontovela.exe ./cmd/ontovela
.\bin\ontovela.exe
```

The default store is in-memory. All API requests require `X-Tenant-ID`; writes
require a prior source binding and an `Idempotency-Key`.

## Option B: PostgreSQL via Docker Compose

```powershell
# From the ONTOVELA core repository
docker compose up -d postgres
$env:ONTOVELA_PG_DSN = "postgres://ontovela:ontovela@localhost:5432/ontovela?sslmode=disable"
cd backend; go run ./cmd/ontovela -migrate
```

## Verify

```powershell
curl -X POST http://localhost:8080/v1/twins -H 'X-Tenant-ID: acme' -H 'Content-Type: application/json' -d '{"id":"robot:WH-17","type_ref":"robot"}'
curl -X POST http://localhost:8080/v1/source-bindings -H 'X-Tenant-ID: acme' -H 'Content-Type: application/json' -d '{"id":"b","source":"sensor:health","property":"health","authority_rank":1,"max_lag_seconds":60}'
curl -X POST http://localhost:8080/v1/assertions -H 'X-Tenant-ID: acme' -H 'Idempotency-Key: k1' -H 'Content-Type: application/json' -d '{"subject_id":"robot:WH-17","property":"health","value":"ready","state_kind":"observed","event_time":"2026-08-11T10:00:00Z","source":"sensor:health","evidence_ref":"e1"}'
curl -H 'X-Tenant-ID: acme' http://localhost:8080/v1/twins/robot:WH-17/state/health
```

The full API contract is in `api/openapi.yaml`; ready-to-run examples are in
`examples/warehouse-robot.md`.
