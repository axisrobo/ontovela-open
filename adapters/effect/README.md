# ONTOVELA Executor Effect Adapter

Maps executor `EffectRecord`s into ONTOVELA state assertions. Defaults to
`observed`; only `reported` may be set explicitly. It never emits
`simulated` or `predicted` state from an effect.

```go
input, err := effect.ToAssertionInput(record)
// append `input` via the normal core write path
```

Run verification:

```powershell
GOWORK=off go test ./...
GOWORK=off go vet ./...
```
