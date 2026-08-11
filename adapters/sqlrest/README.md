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

## PostgreSQL query source example

```go
type pgSource struct{ db *sql.DB }

func (p *pgSource) Fetch(ctx context.Context, after sqlrest.Cursor) ([]sqlrest.Row, sqlrest.Cursor, error) {
    rows, err := p.db.QueryContext(ctx, `
        SELECT id, tenant_id, subject_id, property, value, state_kind,
               event_time, source_ref, evidence_ref
        FROM changed_rows WHERE id > $1 ORDER BY id LIMIT 500`, string(after))
    // map rows to sqlrest.Row, return the last id as the next cursor
}
```

Run verification:

```powershell
GOWORK=off go test ./...
GOWORK=off go vet ./...
```
