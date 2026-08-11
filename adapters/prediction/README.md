# ONTOVELA Prediction Adapter

Self-contained Go reference adapter that turns model predictions into
`predicted` state assertions. It has no API to rewrite predictions as
`observed`, preserving epistemic separation.

```go
predictor, _ := prediction.NewHTTPPredictor("http://model:9000/predict", nil)
result, err := predictor.Predict(ctx, prediction.PredictionRequest{
    SubjectID: "bin:A-01", Property: "demand", ModelRef: "noetivela:demand-v2",
})
input := prediction.ToAssertionInput(result, req, "noetivela:demand-v2", "run/abc", time.Now().UTC())
```

Feed the returned `StateAssertionInput` through the normal ONTOVELA write path
with a tenant source binding and idempotency key.

Run verification:

```powershell
GOWORK=off go test ./...
GOWORK=off go vet ./...
```
