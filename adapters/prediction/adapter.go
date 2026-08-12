// Package prediction is a self-contained ONTOVELA reference adapter that turns
// model predictions into `predicted` state assertions. It has no API to
// rewrite predictions as observed state.
package prediction

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	ontovela "github.com/axisrobo/ONTOVELA-open/sdk/go"
)

var ErrPredictionFailed = errors.New("prediction failed")

// Prediction is a model output normalized for ONTOVELA.
type Prediction struct {
	Value        json.RawMessage `json:"value"`
	Confidence   float64         `json:"confidence"`
	ModelVersion string          `json:"model_version"`
}

// StateAssertionInput mirrors the public StateAssertion write contract. The
// adapter always sets state_kind to "predicted".
type StateAssertionInput struct {
	SubjectID   string          `json:"subject_id"`
	Property    string          `json:"property"`
	Value       json.RawMessage `json:"value"`
	StateKind   string          `json:"state_kind"`
	EventTime   time.Time       `json:"event_time"`
	Source      string          `json:"source"`
	Confidence  *float64        `json:"confidence,omitempty"`
	EvidenceRef string          `json:"evidence_ref"`
}

// Predictor returns a prediction for a request.
type Predictor interface {
	Predict(ctx context.Context, request PredictionRequest) (Prediction, error)
}

// PredictionRequest carries the inputs for one model call.
type PredictionRequest struct {
	SubjectID string         `json:"subject_id"`
	Property  string         `json:"property"`
	ModelRef  string         `json:"model_ref"`
	Inputs    map[string]any `json:"inputs"`
}

// HTTPPredictor calls a model endpoint and parses the normalized prediction.
type HTTPPredictor struct {
	ModelURL string
	HTTP     *http.Client
}

func NewHTTPPredictor(modelURL string, httpClient *http.Client) (*HTTPPredictor, error) {
	parsed, err := url.Parse(modelURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("invalid model URL %q", modelURL)
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &HTTPPredictor{ModelURL: modelURL, HTTP: httpClient}, nil
}

func (p *HTTPPredictor) Predict(ctx context.Context, request PredictionRequest) (Prediction, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return Prediction{}, err
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, p.ModelURL, bytes.NewReader(payload))
	if err != nil {
		return Prediction{}, err
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept", "application/json")
	response, err := p.HTTP.Do(httpRequest)
	if err != nil {
		return Prediction{}, err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return Prediction{}, fmt.Errorf("%w: model returned %d", ErrPredictionFailed, response.StatusCode)
	}
	var prediction Prediction
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&prediction); err != nil {
		return Prediction{}, err
	}
	if !json.Valid(prediction.Value) || prediction.Confidence < 0 || prediction.Confidence > 1 {
		return Prediction{}, fmt.Errorf("%w: invalid value or confidence", ErrPredictionFailed)
	}
	return prediction, nil
}

// AppendTo converts a prediction and appends it to the core as a predicted
// assertion through the public SDK client.
func AppendTo(ctx context.Context, client *ontovela.Client, prediction Prediction, request PredictionRequest, source, evidenceRef string, generatedAt time.Time) error {
	input := ToAssertionInput(prediction, request, source, evidenceRef, generatedAt)
	_, err := client.AppendAssertion(ctx, ontovela.StateAssertionInput{
		SubjectID: input.SubjectID, Property: input.Property, Value: input.Value,
		StateKind: ontovela.StateKind(input.StateKind), EventTime: input.EventTime,
		Source: input.Source, Confidence: input.Confidence, EvidenceRef: input.EvidenceRef,
	}, evidenceRef)
	return err
}

// ToAssertionInput converts a prediction into a predicted StateAssertionInput.
// The generated source and evidence reference identify the model run.
func ToAssertionInput(prediction Prediction, request PredictionRequest, source, evidenceRef string, generatedAt time.Time) StateAssertionInput {
	confidence := prediction.Confidence
	return StateAssertionInput{
		SubjectID:   request.SubjectID,
		Property:    request.Property,
		Value:       prediction.Value,
		StateKind:   "predicted",
		EventTime:   generatedAt,
		Source:      source,
		Confidence:  &confidence,
		EvidenceRef: evidenceRef,
	}
}
