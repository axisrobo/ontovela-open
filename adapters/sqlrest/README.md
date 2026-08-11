# ONTOVELA SQL/REST Adapter

Pull-based reference adapter that polls a SQL query or REST endpoint and ingests
changed rows as state assertions. Cursor advances only after a batch succeeds,
so failures are retried.

```go
import "github.com/axisrobo/ONTOVELA-open/adapters/sqlrest"

poller := &sqlrest.Poller{Source: mySQLSource, Client: client}
cursor, err := poller.PollOnce(ctx, cursor)
```

Implement `RowSource.Fetch` to return normalized rows; tenant scope, idempotency
keys, and state kinds are preserved from rows.

Run verification:

```powershell
GOWORK=off go test ./...
GOWORK=off go vet ./...
```
