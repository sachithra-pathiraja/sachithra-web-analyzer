package service

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"web-analyzer/internal/model"

	"github.com/PuerkitoBio/goquery"
)

func TestGetHTMLVersion_HTML5_NoDoctype(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	html5 := "<!DOCTYPE html><html><head><title>t</title></head><body></body></html>"
	v, err := getHTMLVersion(strings.NewReader(html5), logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "HTML5" {
		t.Fatalf("expected HTML5, got %q", v)
	}

	noDoctype := "<html><head></head><body></body></html>"
	v2, err := getHTMLVersion(strings.NewReader(noDoctype), logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v2 != "No DOCTYPE (Quirks Mode)" {
		t.Fatalf("expected No DOCTYPE (Quirks Mode), got %q", v2)
	}
}

func TestGetTitleAndHeadings(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	html := `<html><head><title>My Title</title></head><body>
    <h1>one</h1><h2>two</h2><h2>twob</h2><h3>three</h3>
    </body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("failed creating doc: %v", err)
	}
	title, headings, err := getTitleAndHeadings(doc, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if title != "My Title" {
		t.Fatalf("unexpected title: %q", title)
	}
	// Expect headings: h1 (1), h2 (2), h3 (1) => 3 entries
	if len(headings) != 3 {
		t.Fatalf("expected 3 heading entries, got %d", len(headings))
	}
	// verify one of them
	var found bool
	for _, h := range headings {
		if h.Level == 2 && h.Count == 2 {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected h2 count 2 in headings: %#v", headings)
	}
	// also ensure returned type matches model.Heading usage (sanity)
	_ = model.Heading{Level: headings[0].Level, Count: headings[0].Count}
}
