# ONTOVELA HTTP/Webhook Adapter

Reference adapter that ingests webhook HTTP payloads into ONTOVELA as
tenant-scoped, idempotent state assertions. It dogfoods the public Go SDK and
never changes state kinds or promotes simulated values.

```go
client, _ := ontovela.NewClient("http://localhost:8080", "acme")
http.Handle("/ingest", &httpwebhook.Server{Client: client})
```

Accepted body:

```json
{
  "tenant_id": "acme",
  "idempotency_key": "webhook-1",
  "subject_id": "robot:WH-17",
  "property": "health",
  "value": "ready",
  "state_kind": "observed",
  "event_time": "2026-08-11T10:00:00Z",
  "source": "webhook:robot",
  "evidence_ref": "webhook:event/1"
}
```

Run verification:

```powershell
GOWORK=off go test ./...
GOWORK=off go vet ./...
```
