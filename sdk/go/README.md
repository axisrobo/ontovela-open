# ONTOVELA Go SDK

```go
client, err := ontovela.NewClient("http://localhost:8080", "acme")
if err != nil { panic(err) }

_, err = client.CreateTwin(ctx, ontovela.TwinInput{
    ID: "robot:WH-17", TypeRef: "robot",
})
```

The SDK has no external dependencies. It covers every v0.1 `/v1` endpoint, sends `X-Tenant-ID` on every request, and requires callers to supply an idempotency key for operational writes.

Run verification from this directory:

```powershell
GOWORK=off go test ./...
GOWORK=off go vet ./...
```
