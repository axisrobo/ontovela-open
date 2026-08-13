# ONTOVELA Live Source-Class Verification

Proves the five stable source classes plus ORCHADYN/PEIRAVELA cross-product
consumption against a real core backed by PostgreSQL.

## Run

1. Start PostgreSQL (local, no Docker).
2. Start the core: `..\scripts\run-live-core.ps1` (applies migrations, serves :8080).
3. In a second terminal:
   ```powershell
   $env:ONTOVELA_LIVE='1'
   $env:GOWORK='off'
   go test ./... -v
   ```

Each test seeds a twin + binding, drives the adapter's real mapping path (e.g.
`effect.ToAssertionInput`, `mqtt.Run`, `httpwebhook.Server`, `opcua.ToAssertionInput`,
`harmovela.Client.FetchAndValidate`), resolves the property, and asserts the
resolved state kind (`observed`) with no `predicted`/`simulated` promotion.
`consumption_test.go` additionally verifies ORCHADYN-style Reality View
consumption and PEIRAVELA-style signed snapshot + sim-to-real branch
comparison over the live core.

Tests skip unless `ONTOVELA_LIVE=1` and the core is reachable.
