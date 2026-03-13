package service

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"web-analyzer/internal/apierror"
	"web-analyzer/internal/model"
)

type errTransport struct{}

func (errTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("round trip error")
}

func TestFetchService_ProcessDocument_Success(t *testing.T) {
	// Server returns a full HTML doc
	html := `<!DOCTYPE html><html><head><title>T</title></head><body>
    <h1>H</h1>
    <a href="/ok">ok</a>
    </body></html>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve same HTML for GET, and OK for HEAD
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	fs := NewFetchService(client, logger, 3)
	doc := &model.Document{URL: srv.URL}

	out, err := fs.ProcessDocument(context.Background(), doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Title != "T" {
		t.Fatalf("expected title T, got %q", out.Title)
	}
	if len(out.Headings) == 0 {
		t.Fatalf("expected headings, got none")
	}
	// Links should return a slice with counts
	if len(out.Links) == 0 {
		t.Fatalf("expected links info, got none")
	}
}

func TestFetchService_ProcessDocument_FetchError(t *testing.T) {
	// client with transport that errors
	client := &http.Client{
		Transport: errTransport{},
		Timeout:   2 * time.Second,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	fs := NewFetchService(client, logger, 1)

	doc := &model.Document{URL: "http://example.invalid/"}

	_, err := fs.ProcessDocument(context.Background(), doc)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	var apiErr *apierror.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected apierror.Error, got %T", err)
	}
	if apiErr.Code != apierror.ErrFetchFailed {
		t.Fatalf("expected ErrFetchFailed, got %q", apiErr.Code)
	}
}
