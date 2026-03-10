package service

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"web-analyzer/internal/apierror"
	"web-analyzer/internal/model"

	"github.com/PuerkitoBio/goquery"
)

func getLinks(
	doc *goquery.Document,
	targetURL string,
	client *http.Client,
	logger *slog.Logger,
	workerCount int,
) ([]model.Link, error) {

	baseURL, err := url.Parse(targetURL)
	if err != nil {

		logger.Error("invalid base url", "url", targetURL, "error", err)

		return nil, apierror.New(
			apierror.ErrInvalidURL,
			"invalid target url",
		)
	}

	type linkCounter struct {
		internal     int
		external     int
		inaccessible int
	}

	stats := linkCounter{}
	seen := make(map[string]struct{})

	var mu sync.Mutex
	var wg sync.WaitGroup

	jobs := make(chan string, workerCount*10)
	if workerCount <= 0 {
		workerCount = 5
	}
	for w := 0; w < workerCount; w++ {
		go func() {
			for urlStr := range jobs {

				resp, err := client.Head(urlStr)

				if err != nil || resp.StatusCode >= 400 {

					logger.Warn("inaccessible link",
						"url", urlStr,
						"error", err,
					)

					mu.Lock()
					stats.inaccessible++
					mu.Unlock()
				}

				if resp != nil {
					resp.Body.Close()
				}

				wg.Done()
			}
		}()
	}

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

		mu.Lock()

		if _, exists := seen[urlStr]; exists {
			mu.Unlock()
			return
		}

		seen[urlStr] = struct{}{}

		if absoluteURL.Host == baseURL.Host {
			stats.internal++
		} else {
			stats.external++
		}

		mu.Unlock()

		wg.Add(1)
		jobs <- urlStr
	})

	close(jobs)
	wg.Wait()

	logger.Info("link analysis completed",
		"internal", stats.internal,
		"external", stats.external,
		"inaccessible", stats.inaccessible,
	)

	return []model.Link{
		{LinkType: "internal", Count: stats.internal},
		{LinkType: "external", Count: stats.external},
		{LinkType: "inaccessible", Count: stats.inaccessible},
	}, nil
}
