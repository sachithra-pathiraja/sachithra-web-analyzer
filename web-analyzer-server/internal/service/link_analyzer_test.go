package service

import (
    "io"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/PuerkitoBio/goquery"
    "web-analyzer/internal/model"
)

func TestGetLinks_CountsInternalExternalInaccessible(t *testing.T) {
    logger := slog.New(slog.NewTextHandler(io.Discard, nil))

    // Server 1: responds OK for HEAD
    srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodHead {
            w.WriteHeader(http.StatusOK)
            return
        }
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("<html></html>"))
    }))
    defer srv1.Close()

    // Server 2: responds 404 for HEAD
    srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method == http.MethodHead {
            w.WriteHeader(http.StatusNotFound)
            return
        }
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("<html></html>"))
    }))
    defer srv2.Close()

    // Build HTML linking to both servers (one internal to base, one external)
    html := `<html><body>
    <a href="` + srv1.URL + `">internal-ok</a>
    <a href="` + srv2.URL + `">external-bad</a>
    </body></html>`

    doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
    if err != nil {
        t.Fatalf("failed creating doc: %v", err)
    }

    // Use a short timeout client
    client := &http.Client{Timeout: 5 * time.Second}

    links, err := getLinks(doc, srv1.URL, client, logger, 5)
    if err != nil {
        t.Fatalf("getLinks returned error: %v", err)
    }

    // Expect 3 link entries returned (Internal, External, Inaccessible)
    var internal, external, inaccessible int
    for _, l := range links {
        switch l.LinkType {
        case "Internal":
            internal = l.Count
        case "External":
            external = l.Count
        case "Inaccessible":
            inaccessible = l.Count
        }
    }

    if internal != 1 {
        t.Fatalf("expected 1 internal link, got %d", internal)
    }
    if external != 1 {
        t.Fatalf("expected 1 external link, got %d", external)
    }
    // external link returned 404 => inaccessible should be 1
    if inaccessible != 1 {
        t.Fatalf("expected 1 inaccessible link, got %d", inaccessible)
    }

    // Sanity: ensure model.Link shape is usable
    _ = model.Link{LinkType: "X", Count: 1}
}