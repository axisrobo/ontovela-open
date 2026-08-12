# Protocol Threat Model

## Threats introduced by external protocols

| Threat | Affected protocols | Mitigation |
| --- | --- | --- |
| Unauthenticated publisher | all push adapters | TLS + protocol auth at the client; ONTOVELA source binding + principal-ref |
| Replay of stale telemetry | all | idempotency keys; event-time skew rejection |
| Spoofed device ID | MQTT, CoAP, LoRaWAN, BLE, CAN | bind source to device identity; principal-ref on bindings |
| Bad-quality data | OPC UA, Modbus, SNMP | adapter quality gates reject before append |
| Order/cursor manipulation | pull adapters | cursor advances only after successful append |
| Data-race / forged frames | CAN, DDS, gRPC | validation at adapter; evidence references kept |
| Cross-tenant injection | all | `base.Payload.TenantID` must match the SDK tenant |
| Payload smuggling (state kind) | all | only six state kinds; `simulated`/`predicted` never promoted |

## Core invariants regardless of protocol

- Writes require a tenant source binding and evidence reference.
- A revoked binding immediately stops writes from that source.
- Archived twins reject new claims.
- Conflict records make multi-source disagreement explicit.
- Snapshot digests and Reality View policy pinning prevent substitution.
