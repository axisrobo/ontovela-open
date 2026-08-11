# ONTOVELA OPC UA Adapter

Maps OPC UA node reads into observed assertions. Poor-quality reads are
rejected so bad data never becomes state.

```go
input, err := opcua.ToAssertionInput(node)
// append via the normal core write path
```

Run verification:

```powershell
GOWORK=off go test ./...
GOWORK=off go vet ./...
```
