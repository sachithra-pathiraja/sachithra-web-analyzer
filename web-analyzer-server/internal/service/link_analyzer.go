package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"web-analyzer/internal/apierror"
	"web-analyzer/internal/model"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/sync/errgroup"
)

func getLinks(
	doc *goquery.Document,
	targetURL string,
	client *http.Client,
	logger *slog.Logger,
	workerCount int,
) ([]model.Link, error) {

	// 1) Normalize workerCount BEFORE using it anywhere
	if workerCount <= 0 {
		workerCount = 5
	}

	baseURL, err := url.Parse(targetURL)
	if err != nil {
		logger.Error("invalid base url", "url", targetURL, "error", err)
		return nil, apierror.New(apierror.ErrInvalidURL, "invalid target url", targetURL)
	}

	type linkCounter struct {
		internal     int
		external     int
		inaccessible int
	}

	var (
		stats linkCounter
		mu    sync.Mutex
	)

	// 2) Extract + dedupe first (single goroutine => no races, no wg/channel hazards)
	seen := make(map[string]struct{})
	var urls []string

	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		parsedHref, err := url.Parse(href)
		if err != nil {
			logger.Warn("invalid href skipped", "href", href)
			return
		}

		if parsedHref.Scheme == "mailto" ||
			parsedHref.Scheme == "javascript" ||
			parsedHref.Scheme == "tel" ||
			strings.HasPrefix(href, "#") {
			return
		}

		absoluteURL := baseURL.ResolveReference(parsedHref)
		urlStr := absoluteURL.String()

		if _, ok := seen[urlStr]; ok {
			return
		}
		seen[urlStr] = struct{}{}

		if absoluteURL.Host == baseURL.Host {
			stats.internal++
		} else {
			stats.external++
		}

		urls = append(urls, urlStr)
	})

	// 3) Bounded concurrency with errgroup + semaphore
	// Add a request timeout to prevent hanging workers
	g, ctx := errgroup.WithContext(context.Background())

	sem := make(chan struct{}, workerCount)

	for _, u := range urls {
		u := u // capture
		g.Go(func() error {
			// Acquire semaphore slot or exit if context cancelled
			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return ctx.Err()
			}
			defer func() { <-sem }()

			// Use per-request timeout
			reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			req, err := http.NewRequestWithContext(reqCtx, http.MethodHead, u, nil)
			if err != nil {
				// treat as inaccessible
				apiErr := apierror.New(apierror.ErrRequestCreation, fmt.Sprintf("failed creating request for %s: %v", u, err), targetURL)
				logger.Warn("request creation failed", "url", u, "apierror", apiErr)
				mu.Lock()
				stats.inaccessible++
				mu.Unlock()
				// non-fatal: continue scanning other links
				return nil
			}

			resp, err := client.Do(req)
			if err != nil {
				// treat as inaccessible
				apiErr := apierror.New(apierror.ErrRequestFailed, fmt.Sprintf("request failed for %s: %v", u, err), targetURL)
				logger.Warn("inaccessible link", "url", u, "apierror", apiErr)
				mu.Lock()
				stats.inaccessible++
				mu.Unlock()
				return nil
			}

			if resp.StatusCode >= 400 {
				apiErr := apierror.New(apierror.ErrInaccessibleLink, fmt.Sprintf("status %d for %s", resp.StatusCode, u), targetURL)
				logger.Warn("inaccessible link", "url", u, "apierror", apiErr)
				mu.Lock()
				stats.inaccessible++
				mu.Unlock()
			}

			if resp.Body != nil {
				resp.Body.Close()
			}
			return nil
		})
	}

	// Note: we don't *need* to fail the whole operation on request errors,
	// so each goroutine returns nil even if the link is inaccessible.
	if err := g.Wait(); err != nil {
		// Map context/cancel/timeout to a timeout-specific apierror
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			logger.Error("link analysis cancelled or timed out", "error", err)
			return nil, apierror.New(apierror.ErrRequestTimeout, "link analysis timed out or cancelled", targetURL)
		}
		// Any other errgroup error => wrap and return
		logger.Error("link analysis failed", "error", err)
		return nil, apierror.New(apierror.ErrRequestFailed, err.Error(), targetURL)
	}

	logger.Info("link analysis completed",
		"internal", stats.internal,
		"external", stats.external,
		"inaccessible", stats.inaccessible,
	)

	return []model.Link{
		{LinkType: "Internal", Count: stats.internal},
		{LinkType: "External", Count: stats.external},
		{LinkType: "Inaccessible", Count: stats.inaccessible},
	}, nil
}

func statusCode(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}
