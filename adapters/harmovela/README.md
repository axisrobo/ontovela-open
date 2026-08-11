# ONTOVELA Harmovela Evidence Adapter

Self-contained Go reference adapter that resolves and validates ONTOVELA
`harmovela:event/<id>` evidence references against the public Harmovela event
contract. It does not import Harmovela implementation internals.

```go
client, _ := harmovela.NewClient("http://harmovela:8080", nil)
record, err := client.FetchAndValidate(ctx, "harmovela:event/9f2...")
```

Validation requires `spec_version == "0.2"`, non-empty `id`, `type`, `source`,
an RFC3339 `created_at`, and a present `payload`. Missing or malformed events
fail explicitly so unverifiable evidence never supports a resolved state.

Run verification:

```powershell
GOWORK=off go test ./...
GOWORK=off go vet ./...
```
