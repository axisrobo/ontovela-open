package prediction

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/axisrobo/ONTOVELA-open/adapters/base/testutil"
	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

// TestPredictionToAssertionIntegration drives model -> adapter -> core append.
func TestPredictionToAssertionIntegration(t *testing.T) {
	model := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Prediction{Value: json.RawMessage(`3.2`), Confidence: 0.84, ModelVersion: "demand-v2"})
	}))
	defer model.Close()

	core := testutil.NewFakeCore()
	defer core.Close()

	predictor, err := NewHTTPPredictor(model.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	client, err := ontovela.NewClient(core.URL(), "acme")
	if err != nil {
		t.Fatal(err)
	}
	request := PredictionRequest{SubjectID: "bin:A-01", Property: "demand", ModelRef: "noetivela:demand-v2", Inputs: map[string]any{"sku": "X1"}}
	prediction, err := predictor.Predict(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	input := ToAssertionInput(prediction, request, "noetivela:demand-v2", "run/abc", time.Now().UTC())
	if input.StateKind != "predicted" {
		t.Fatalf("state kind = %q", input.StateKind)
	}
	// The predicted assertion reaches the core without promoting to observed.
	if err := AppendTo(context.Background(), client, prediction, request, "noetivela:demand-v2", "run/abc", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if core.Count() != 1 {
		t.Fatalf("appends = %d", core.Count())
	}
}
