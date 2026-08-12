# Stable Source Classes

The Phase 2 exit gate requires five stable source classes. Each class maps to
one or more ONTOVELA adapters and a reference `SourceBinding` shape.

| Class | Example sources | Adapters | Binding notes |
| --- | --- | --- | --- |
| Executor effects | KINETOVELA, PRAXOVELA, RHEOVELA | `effect` | `state_kind` observed/reported; evidence ref is the effect ID |
| Event streams | Kafka, NATS, AMQP, MQTT, WebSocket, gRPC | `stream`, `mqtt`, `amqp`, `amqp1`, `grpc`, `websocket`, `sse`, `stomp`, `zeromq`, `redis`, `graphql`, `mqttsn` | tenant + idempotency per message; broker-neutral `Source` contract |
| Enterprise SQL/REST | ERP, WMS, CMDB | `sqlrest`, `csv`, `longpoll`, `httpwebhook` | pull cursor or push payload; mapping to twin packs |
| IoT / industrial telemetry | OPC UA, Modbus, EtherNet/IP, PROFINET, SNMP, BACnet, CAN, CoAP, LoRaWAN, BLE, DDS | `opcua`, `modbus`, `ethernetip`, `profinet`, `snmp`, `bacnet`, `can`, `coap`, `lorawan`, `ble`, `dds` | quality gates (e.g., OPC UA good quality) enforced per adapter |
| Evidence references | Harmovela | `harmovela` | `evidence_ref` validated against the event contract |

## Reference binding

```json
{
  "id": "wms-onhand",
  "source": "wms:bin-1",
  "property": "available",
  "authority_rank": 10,
  "max_lag_seconds": 300,
  "principal_ref": "principal:wms-1"
}
```

Every class preserves tenant scope, idempotency, source, evidence, and state
kind; none can promote `simulated` or `predicted` to `observed`.
