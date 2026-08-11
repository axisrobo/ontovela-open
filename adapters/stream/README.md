# ONTOVELA Stream Ingest Adapter

Broker-neutral reference adapter for Kafka, NATS, and similar message streams.
Kafka and NATS client libraries implement the `Consumer` interface without
adding broker code to ONTOVELA.

```go
import "github.com/axisrobo/ONTOVELA-open/adapters/stream"

client, _ := ontovela.NewClient("http://localhost:8080", "acme")
err := stream.Run(ctx, myKafkaConsumer, client)
```

Each message body is a JSON `IngestRequest` with tenant, idempotency key,
state kind, value, source, and evidence reference. Cross-tenant messages and
invalid state kinds are rejected before reaching the core.

Run verification:

```powershell
GOWORK=off go test ./...
GOWORK=off go vet ./...
```
