# ONTOVELA Edge Spool Adapter

Offline-tolerant edge adapter that buffers state assertions locally and replays
them in order when connectivity returns.

```go
import "github.com/axisrobo/ONTOVELA-open/adapters/edge"

spool := edge.New()
sequence := spool.Append(item) // offline-safe
err := spool.Flush(ctx, client) // replays in order, then clears
```

Buffered items carry idempotency keys, so replay never duplicates state. A
failed flush stops and retains the remaining items.

Run verification:

```powershell
GOWORK=off go test ./...
GOWORK=off go vet ./...
```
