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

## Kafka consumer example

```go
type kafkaSource struct{ reader *kafka.Reader } // kafka-go

func (k *kafkaSource) Consume(ctx context.Context, handler func(stream.Message) error) error {
    for {
        message, err := k.reader.ReadMessage(ctx)
        if err != nil { return err }
        if err := handler(stream.Message{Body: message.Value, Topic: message.Topic, Partition: message.Partition, Offset: message.Offset}); err != nil {
            return err
        }
    }
}
```

## NATS consumer example

```go
type natsSource struct{ conn *nats.Conn }

func (n *natsSource) Consume(ctx context.Context, handler func(stream.Message) error) error {
    subject := "ontovela.state.>"
    sub, err := n.conn.Subscribe(subject, func(message *nats.Msg) {
        _ = handler(stream.Message{Body: message.Data})
    })
    if err != nil { return err }
    <-ctx.Done()
    _ = sub.Unsubscribe()
    return nil
}
```

Run verification:

```powershell
GOWORK=off go test ./...
GOWORK=off go vet ./...
```
