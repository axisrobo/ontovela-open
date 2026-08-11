package prediction

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestToAssertionInputIsAlwaysPredicted(t *testing.T) {
	when := time.Now().UTC()
	prediction := Prediction{Value: json.RawMessage(`3.2`), Confidence: 0.84, ModelVersion: "demand-v2"}
	request := PredictionRequest{SubjectID: "bin:A-01", Property: "demand", ModelRef: "noetivela:demand-v2"}
	input := ToAssertionInput(prediction, request, "noetivela:demand-v2", "run/abc", when)
	if input.StateKind != "predicted" {
		t.Fatalf("state kind = %q, want predicted", input.StateKind)
	}
	if input.Confidence == nil || *input.Confidence != 0.84 {
		t.Fatalf("confidence = %v", input.Confidence)
	}
	if string(input.Value) != `3.2` {
		t.Fatalf("value = %s", input.Value)
	}
}

func TestHTTPPredictorParsesModelOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Prediction{Value: json.RawMessage(`5`), Confidence: 0.9, ModelVersion: "v1"})
	}))
	defer server.Close()
	predictor, err := NewHTTPPredictor(server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	prediction, err := predictor.Predict(context.Background(), PredictionRequest{SubjectID: "bin:A-01", Property: "demand"})
	if err != nil {
		t.Fatal(err)
	}
	if string(prediction.Value) != `5` || prediction.Confidence != 0.9 {
		t.Fatalf("prediction = %#v", prediction)
	}
}

func TestHTTPPredictorRejectsInvalidConfidence(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Prediction{Value: json.RawMessage(`5`), Confidence: 1.5})
	}))
	defer server.Close()
	predictor, err := NewHTTPPredictor(server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := predictor.Predict(context.Background(), PredictionRequest{}); err == nil {
		t.Fatal("expected invalid confidence rejection")
	}
}
