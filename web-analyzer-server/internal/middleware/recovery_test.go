package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"log/slog"
)

func TestRecovery_RecoversPanicAndReturns500(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	h := Recovery(logger)(panicHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	res := w.Result()
	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d, body: %s", res.StatusCode, string(body))
	}

	logOut := buf.String()
	if !strings.Contains(logOut, "panic recovered") {
		t.Fatalf("expected log to contain 'panic recovered', got: %q", logOut)
	}
	if !strings.Contains(logOut, "boom") {
		t.Fatalf("expected log to contain panic value 'boom', got: %q", logOut)
	}
}

func TestRecovery_PassesThroughWhenNoPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	h := Recovery(logger)(okHandler)

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	res := w.Result()
	body, _ := io.ReadAll(res.Body)

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}
	if string(body) != "ok" {
		t.Fatalf("expected body 'ok', got %q", string(body))
	}

	logOut := buf.String()
	if strings.Contains(logOut, "panic recovered") {
		t.Fatalf("did not expect panic log when no panic occurred, got: %q", logOut)
	}
}
