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

func TestLogging_WritesLogAndPassesThroughResponse(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	h := Logging(logger)(final)

	req := httptest.NewRequest(http.MethodGet, "/testpath", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	// response preserved
	res := w.Result()
	body, _ := io.ReadAll(res.Body)
	if string(body) != "ok" {
		t.Fatalf("expected response body 'ok', got %q", string(body))
	}

	out := buf.String()
	if !strings.Contains(out, "request completed") {
		t.Fatalf("log output missing message: %q", out)
	}
	if !strings.Contains(out, "method=GET") {
		t.Fatalf("log output missing method: %q", out)
	}
	if !strings.Contains(out, "path=/testpath") {
		t.Fatalf("log output missing path: %q", out)
	}
	if !strings.Contains(out, "remote_addr=1.2.3.4:1234") {
		t.Fatalf("log output missing remote_addr: %q", out)
	}
}

func TestLogging_IncludesDurationField(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// noop
	})
	h := Logging(logger)(final)

	req := httptest.NewRequest(http.MethodPost, "/d", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	out := buf.String()
	if !strings.Contains(out, "duration=") {
		t.Fatalf("log output missing duration field: %q", out)
	}
}
