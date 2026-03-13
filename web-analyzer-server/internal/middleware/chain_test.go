package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChain_AppliesMiddlewaresInOrder(t *testing.T) {
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("handler;"))
	})

	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("m1-start;"))
			next.ServeHTTP(w, r)
			_, _ = w.Write([]byte("m1-end;"))
		})
	}

	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("m2-start;"))
			next.ServeHTTP(w, r)
			_, _ = w.Write([]byte("m2-end;"))
		})
	}

	h := Chain(final, m1, m2)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	res := w.Result()
	body, _ := io.ReadAll(res.Body)
	got := strings.TrimSpace(string(body))

	expected := "m1-start;m2-start;handler;m2-end;m1-end;"
	if got != expected {
		t.Fatalf("unexpected chain output:\nexpected: %q\n got: %q", expected, got)
	}
}

func TestChain_NoMiddlewares_ReturnsHandlerUnchanged(t *testing.T) {
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	h := Chain(final) // no middlewares

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	res := w.Result()
	body, _ := io.ReadAll(res.Body)
	got := string(body)

	if got != "ok" {
		t.Fatalf("expected body 'ok', got %q", got)
	}
}
