# Enterprise Incident Response Reference

Tie business orders, service dependencies, process instances, agent execution,
regional state, and personnel availability into one incident Reality View.

## Flow

1. **Write state**: KINETOVELA writes agent execution projections, enterprise
   systems write order and service health, sensors write regional state, and
   process twins receive stage updates — all through tenant source bindings.
2. **Resolve impact**: `GET /v1/twins/{incident}/impact` surfaces which orders,
   services, and commitments depend on the affected component.
3. **Decision view**: a Reality View over `{service.health, order.stage,
   region.state, agent.status}` with purpose-bound `max_age_seconds` returns
   `ready`, `stale`, `unknown`, or `conflicted`.
4. **Reconstruct**: `POST /v1/twins/{incident}/snapshots` pins the incident
   world; `GET /v1/snapshots/{id}/verify` proves the digest so post-incident
   review sees exactly what the operators saw.

## Reality Integrity notes

- `simulated` severity scenarios never enter the incident view.
- Conflicting telemetry opens a `conflict_records` entry instead of a silent
  winner.
- Every state change is reconstructable at the event time it was observed.
