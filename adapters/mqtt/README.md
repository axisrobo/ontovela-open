# ONTOVELA MQTT Adapter

Broker-neutral MQTT-style topic ingestion. An MQTT client implements
`TopicSource`; the adapter maps topic messages to state assertions.

```go
err := mqtt.Run(ctx, myMQTTClient, []string{"ontovela/state/+/+"}, client)
```

Message payloads are JSON `Payload` records with tenant scope, idempotency,
state kind, source, and evidence. Cross-tenant messages are rejected.

Run verification:

```powershell
GOWORK=off go test ./...
GOWORK=off go vet ./...
```
