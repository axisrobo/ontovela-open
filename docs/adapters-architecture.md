# Adapters Architecture

Every protocol adapter follows one pattern:

1. A transport-specific `Message` normalizes the incoming unit (frame, delivery,
   event, record).
2. A `Source` interface isolates the protocol client; a `MemorySource` makes it
   testable offline.
3. `Run` unmarshals `Message.Body` into the shared `base.Payload`, validates
   it, and appends through the public Go SDK.

The shared `base` module enforces the integrity invariants once: tenant match,
idempotency key, subject/property/source/evidence presence, valid state kind,
and valid JSON value. Adapters never implement their own payload semantics.

## Layering

```text
Protocol client (MQTT, Kafka, OPC UA, ...)   -- external, pluggable
      |  Source interface
      v
Adapter module (adapters/<protocol>)          -- normalization + mapping
      |  base.Payload + base.Append
      v
ONTOVELA core (append / resolve / subscribe) -- authoritative ledger
```

## Adding an adapter

1. Add the module under `adapters/<name>` with `require base` and `sdk/go`.
2. Implement `Message`, `Source`, and `Run`.
3. Add a `MemorySource`-based test using `testutil.FakeCore`.
4. Register the adapter in `README.md` and the CI matrix.
