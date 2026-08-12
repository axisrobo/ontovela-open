# Adapter Conformance Matrix

Every adapter must satisfy the same integrity contract:

| Requirement | Guarantee |
| --- | --- |
| Tenant scope | `base.Payload.TenantID` must equal the SDK client tenant; otherwise rejected |
| Idempotency | a non-empty `idempotency_key` is required; replay is safe |
| Source + evidence | `source` and `evidence_ref` are required on every payload |
| State kind | only the six ONTOVELA state kinds; never promoted |
| Value validity | `value` must be valid JSON within size bounds |
| Timestamps | `event_time` present and within skew; system time assigned by the core |

| Adapter | Tenant check | Idempotency | Quality gate | Cursor/order |
| --- | --- | --- | --- | --- |
| `amqp`, `amqp1`, `stomp`, `zeromq`, `grpc`, `websocket`, `sse`, `graphql`, `dds`, `mqttsn` | base | payload | none | per-message |
| `modbus`, `can`, `bacnet`, `ethernetip`, `profinet`, `snmp`, `lorawan`, `ble`, `coap` | base | payload | adapter-specific | per-frame |
| `opcua` | base | payload | quality good | per-read |
| `stream`, `mqtt`, `redis` | base | payload | none | per-message |
| `sqlrest`, `longpoll`, `csv` | base | payload | cursor | after success |
| `effect` | base | payload | observed/reported only | after effect |
| `prediction` | base | payload | model confidence | after model run |
| `harmovela` | n/a (evidence) | n/a | envelope validity | pull |

All adapters are validated in CI via `go test ./...` per module and a
`MemorySource` + `testutil.FakeCore` contract test.
