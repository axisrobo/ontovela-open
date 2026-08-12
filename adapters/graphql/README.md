# ONTOVELA graphql Adapter

GraphQL subscriptions to ONTOVELA assertions. Implements the broker-neutral pattern: a client library provides a Source, and Run maps message bodies to assertions via the shared base payload contract.

Run verification:
```powershell
GOWORK=off go test ./...
```
