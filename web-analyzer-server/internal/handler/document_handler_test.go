package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"web-analyzer/internal/apierror"
	"web-analyzer/internal/model"
)

type mockProcessorOK struct{}

func (mockProcessorOK) ProcessDocument(ctx context.Context, d *model.Document) (*model.Document, error) {
	// return input doc populated
	out := *d
	out.Title = "ok"
	return &out, nil
}

type mockProcessorAPIErr struct {
	code string
}

func (m mockProcessorAPIErr) ProcessDocument(ctx context.Context, d *model.Document) (*model.Document, error) {
	return nil, apierror.New(m.code, "boom", d.URL)
}

func TestAnalyzerHandler_Analyze_Success(t *testing.T) {
	h := NewAnalyzerHandler(mockProcessorOK{})
	body, _ := json.Marshal(model.Document{URL: "http://example/"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Analyze(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var got model.Document
	if err := json.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if got.Title != "ok" {
		t.Fatalf("expected title ok, got %q", got.Title)
	}
}

func TestAnalyzerHandler_Analyze_ApiErrorMapping(t *testing.T) {
	// invalid url mapping -> 400
	h := NewAnalyzerHandler(mockProcessorAPIErr{code: apierror.ErrInvalidURL})
	body, _ := json.Marshal(model.Document{URL: "x"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Analyze(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
	var resp map[string]string
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["code"] != apierror.ErrInvalidURL {
		t.Fatalf("expected code %q, got %q", apierror.ErrInvalidURL, resp["code"])
	}

	// timeout mapping -> 504
	h2 := NewAnalyzerHandler(mockProcessorAPIErr{code: apierror.ErrRequestTimeout})
	body2, _ := json.Marshal(model.Document{URL: "x"})
	req2 := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body2))
	w2 := httptest.NewRecorder()

	h2.Analyze(w2, req2)

	res2 := w2.Result()
	if res2.StatusCode != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d", res2.StatusCode)
	}
}
