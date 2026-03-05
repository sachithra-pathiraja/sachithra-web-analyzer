package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"web-analyzer/internal/model"
	"web-analyzer/internal/repository"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

type FetchService struct {
	repo   *repository.DocumentRepository
	client *http.Client
}

func NewFetchService(r *repository.DocumentRepository) *FetchService {
	return &FetchService{
		repo: r,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *FetchService) ProcessDocument(ctx context.Context, doc *model.Document) (*model.Document, error) {
	log.Printf("Processing document: %+v", doc)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, doc.URL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Read the response body once and reuse readers for parsing
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	rdr := bytes.NewReader(bodyBytes)
	doc.HTMLVersion = getHTMLVersion(rdr)

	// reuse the reader for title extraction
	if _, err := rdr.Seek(0, io.SeekStart); err != nil {
		// if seek fails, try creating a new reader
		rdr = bytes.NewReader(bodyBytes)
	}
	docFromReader, err := goquery.NewDocumentFromReader(rdr)
	if err != nil {
		return nil, err
	}
	doc.Title, doc.Headings = getTitleAndHeadings(docFromReader)
	doc.Links = getLinks(docFromReader, doc.URL, s.client)
	doc.HasLoginForm = getHasLogin(docFromReader)

	return doc, nil
}

func getHTMLVersion(r io.Reader) string {
	tokenizer := html.NewTokenizer(r)

	for {
		tt := tokenizer.Next()

		if tt == html.DoctypeToken {
			token := tokenizer.Token()
			raw := token.String()
			log.Println("RAW DOCTYPE:", raw)
			// HTML5
			if token.Data == "html" && len(token.Attr) == 0 {
				return "HTML5"
			}

			// XHTML
			if strings.Contains(raw, "XHTML 1.0") {
				return "XHTML 1.0"
			}

			// HTML 4
			if strings.Contains(raw, "HTML 4.01") {
				return "HTML 4.01"
			}

			return "Legacy / Unknown DOCTYPE"
		}

		if tt == html.ErrorToken {
			break
		}
	}

	return "No DOCTYPE (Quirks Mode)"
}

func getTitleAndHeadings(doc *goquery.Document) (string, []model.Heading) {
	var headings []model.Heading

	for i := 1; i <= 6; i++ {
		tag := fmt.Sprintf("h%d", i)

		count := doc.Find(tag).Length()

		if count > 0 {
			headings = append(headings, model.Heading{
				Level: i,
				Count: count,
			})
		}
	}
	return doc.Find("title").Text(), headings
}

func getLinks(doc *goquery.Document, targetURL string, client *http.Client) []model.Link {
	baseURL, err := url.Parse(targetURL)
	if err != nil {
		return nil
	}

	type linkCounter struct {
		internal     int
		external     int
		inaccessible int
	}
	stats := linkCounter{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	jobs := make(chan string)

	// ---- Start Worker Pool (5 workers) ----
	workerCount := 5
	for w := 0; w < workerCount; w++ {
		go func() {
			for urlStr := range jobs {
				resp, err := client.Head(urlStr)
				if err != nil || resp.StatusCode >= 400 {
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

	// ---- Produce Jobs ----
	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		parsedHref, err := url.Parse(href)
		if err != nil {
			return
		}

		if parsedHref.Scheme == "mailto" ||
			parsedHref.Scheme == "javascript" ||
			parsedHref.Scheme == "tel" ||
			strings.HasPrefix(href, "#") {
			return
		}

		absoluteURL := baseURL.ResolveReference(parsedHref)

		mu.Lock()
		if absoluteURL.Host == baseURL.Host {
			stats.internal++
		} else {
			stats.external++
		}
		mu.Unlock()

		wg.Add(1)
		jobs <- absoluteURL.String()
	})

	close(jobs)
	wg.Wait()

	return []model.Link{
		{LinkType: "internal", Count: stats.internal},
		{LinkType: "external", Count: stats.external},
		{LinkType: "inaccessible", Count: stats.inaccessible},
	}
}

func getHasLogin(doc *goquery.Document) bool {
	return doc.Find("input[type='password']").Length() > 0
}
