# Multi-Protocol Ingestion

Different source classes feed the same reality through protocol adapters.

| Source class | Protocol | Adapter | Example source |
| --- | --- | --- | --- |
| Executor effects | EffectRecord | `effect` | `kinetovela:robot-1` |
| Event stream | Kafka/NATS | `stream` | `kafka:bin-state` |
| Industrial PLC | OPC UA | `opcua` | `opcua:line-1` |
| Telemetry | MQTT | `mqtt` | `mqtt:robot/WH-17` |
| ERP | SQL/REST | `sqlrest` | `erp:bin` |

Each adapter maps its payload through `base.Payload`:

```json
{
  "tenant_id": "acme",
  "idempotency_key": "<protocol>-<sequence>",
  "subject_id": "robot:WH-17",
  "property": "health",
  "value": "ready",
  "state_kind": "observed",
  "event_time": "2026-08-12T10:00:00Z",
  "source": "<protocol>:<instance>",
  "evidence_ref": "<protocol>/<id>"
}
```

The core resolves authority and conflict across all protocols uniformly.
