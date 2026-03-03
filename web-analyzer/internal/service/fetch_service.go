package service

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"web-analyzer/internal/model"
	"web-analyzer/internal/repository"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

type FetchService struct {
	repo *repository.DocumentRepository
}

func NewFetchService(r *repository.DocumentRepository) *FetchService {
	return &FetchService{repo: r}
}

func (s *FetchService) ProcessDocument(doc *model.Document) (*model.Document, error) {
	log.Printf("Processing document: %+v", doc)
	resp, err := http.Get(doc.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body once and reuse readers for parsing
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	doc.Body = string(bodyBytes)

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
	doc.Links = getLinks(docFromReader, doc.URL)

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

func getLinks(doc *goquery.Document, targetURL string) []model.Link {
	baseURL, _ := url.Parse(targetURL)

	internal := 0
	external := 0
	inaccessible := 0

	client := http.Client{Timeout: 5 * time.Second}

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")

		parsedHref, err := url.Parse(href)
		if err != nil {
			return
		}

		// Ignore mailto, javascript, tel
		if parsedHref.Scheme == "mailto" ||
			parsedHref.Scheme == "javascript" ||
			parsedHref.Scheme == "tel" {
			return
		}

		// Convert to absolute URL
		absoluteURL := baseURL.ResolveReference(parsedHref)

		if absoluteURL.Host == baseURL.Host {
			internal++
			log.Printf("Processing internal link: %s", absoluteURL.String())
		} else {
			external++
			log.Printf("Processing external link: %s", absoluteURL.String())
		}

		resp, err := client.Head(absoluteURL.String())
		if err != nil || resp.StatusCode >= 400 {
			log.Printf("Processing inaccessible link: %s", absoluteURL.String())
			inaccessible++
		}

		if resp != nil {
			resp.Body.Close()
		}
	})

	return []model.Link{
		{LinkType: "internal", Count: internal},
		{LinkType: "external", Count: external},
		{LinkType: "inaccessible", Count: inaccessible},
	}
}
