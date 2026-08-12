# Adapters Deployment Reference

## Model

Adapters are standalone Go processes or in-process workers. They connect a
protocol client to the ONTOVELA core via the public SDK and the shared `base`
payload contract.

## Guidance

- Run one adapter process per source class so a broker or protocol failure is
  contained.
- Adapters are stateless: buffering (offline replay) is handled by the `edge`
  spool, not by the transport adapter.
- Use the protocol client's reconnect/retry behavior; the SDK retries transient
  429/5xx on the core side.
- Preserve idempotency keys end to end so replays never duplicate state.
- For pull adapters (`sqlrest`, `snmp`, `csv`, `longpoll`, `harmovela`,
  `prediction`), advance the cursor only after a successful append.
- Enforce quality gates at the adapter (e.g., OPC UA `good` quality, Modbus
  unit reachability) so bad telemetry never becomes state.

## Observability

- Adapter logs carry the SDK request ID from the `Idempotency-Replayed` and
  trace headers.
- Core `/metrics` and `/healthz` reflect core readiness; adapters report their
  own liveness to the orchestrator.
