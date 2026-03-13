package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"web-analyzer-client/internal/apierror"
	"web-analyzer-client/internal/model"
)

func TestCallAnalyzer_Success(t *testing.T) {
	resp := model.Response{
		URL:   "http://example.com",
		Title: "Example",
	}
	respBytes, _ := json.Marshal(resp)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)
	}))
	defer srv.Close()

	svc := NewAnalyzerService(srv.URL)
	out, err := svc.CallAnalyzer("http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Title != "Example" {
		t.Fatalf("expected title 'Example', got %q", out.Title)
	}
}

func TestCallAnalyzer_ServerError(t *testing.T) {
	apiErr := model.APIError{Code: "bad_url", Message: "Invalid URL"}
	errBytes, _ := json.Marshal(apiErr)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errBytes)
	}))
	defer srv.Close()

	svc := NewAnalyzerService(srv.URL)
	_, err := svc.CallAnalyzer("bad-url")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var ae *apierror.Error
	if !errors.As(err, &ae) {
		t.Fatalf("expected apierror.Error, got %T", err)
	}
	if ae.Code != apierror.ErrServerError {
		t.Fatalf("expected ErrServerError, got %q", ae.Code)
	}
}

func TestCallAnalyzer_InvalidResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	svc := NewAnalyzerService(srv.URL)
	_, err := svc.CallAnalyzer("bad-url")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var ae *apierror.Error
	if !errors.As(err, &ae) {
		t.Fatalf("expected apierror.Error, got %T", err)
	}
	if ae.Code != apierror.ErrInvalidResponse {
		t.Fatalf("expected ErrInvalidResponse, got %q", ae.Code)
	}
}

func TestCallAnalyzer_UnmarshalFailed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{bad json"))
	}))
	defer srv.Close()

	svc := NewAnalyzerService(srv.URL)
	_, err := svc.CallAnalyzer("http://example.com")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var ae *apierror.Error
	if !errors.As(err, &ae) {
		t.Fatalf("expected apierror.Error, got %T", err)
	}
	if ae.Code != apierror.ErrUnmarshalFailed {
		t.Fatalf("expected ErrUnmarshalFailed, got %q", ae.Code)
	}
}

func TestCallAnalyzer_RequestFailed(t *testing.T) {
	// Use an invalid URL to trigger request failure
	svc := NewAnalyzerService("http://localhost:0")
	_, err := svc.CallAnalyzer("http://example.com")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var ae *apierror.Error
	if !errors.As(err, &ae) {
		t.Fatalf("expected apierror.Error, got %T", err)
	}
	if ae.Code != apierror.ErrRequestFailed {
		t.Fatalf("expected ErrRequestFailed, got %q", ae.Code)
	}
}
