# Live Source-Class Verification

Five stable source classes verified against a real core + PostgreSQL (2026-08-13).

| Source class | Adapter | Live proof |
| --- | --- | --- |
| Executor effects | `effect` | EffectRecord → assertion via real SDK → resolved `observed` |
| Event streams | `mqtt` | broker-shaped message via `Run` path → resolved `observed` |
| Enterprise SQL/REST | `httpwebhook` | `POST /ingest`-equivalent mapping → resolved `observed` |
| Industrial telemetry | `opcua` | good-quality node read → resolved `observed` |
| Evidence references | `harmovela` | envelope validated against local HTTP evidence → resolved `observed` |

Run: `scripts/run-live-core.ps1`, then `$env:ONTOVELA_LIVE='1'; go test ./... -v` in `live/`.

Each test asserts tenant scope, idempotency, source binding, evidence reference,
and that no `predicted`/`simulated` value is promoted to `observed`.

## Running against a clean database

The live suite must run against a PostgreSQL database that has **not** been
subjected to EE `pgtenant.ApplyRLS`. RLS policies require
`current_setting('ontovela.tenant_id')` to match every row, which the core
HTTP process does not set; writing into an RLS-enabled database returns 500.
Create a dedicated database (e.g. `ontovela_live`) for live verification:

```sql
CREATE DATABASE ontovela_live OWNER ontovela;
```

Then start the core against it:

```powershell
.\scripts\run-live-core.ps1 -PgDsn "postgres://ontovela:ontovela@localhost:5432/ontovela_live?sslmode=disable"
```

2026-08-13 result: 5/5 `TestLive*` source-class tests PASS against a real core backed by
PostgreSQL on `ontovela_live`; the default `ontovela` database returned 500 on
writes because it carries EE RLS policies.

## Cross-product consumption (2026-08-13)

`live/consumption_test.go` adds live ORCHADYN/PEIRAVELA consumption over the
same real core + PostgreSQL:

| Consumer | Test | Proof |
| --- | --- | --- |
| ORCHADYN | `TestLiveOrchadynRealityViewConsumption` | Reality View from an observed sensor binding → status `ready`, item acceptable, state kind stays `observed` |
| PEIRAVELA | `TestLivePeiravelaSignedSnapshotConsumption` | signed snapshot created + verified; simulated branch diverges (`delta=diverges`) with real state never promoted from `observed` |

Note: `source_bindings` enforces `UNIQUE (tenant_id, source_ref, property)`, so
each test uses a distinct source (and the simulated branch gets its own
binding).
