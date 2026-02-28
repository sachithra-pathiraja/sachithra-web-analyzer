package service

import (
	"io"
	"log"
	"net/http"
	"strings"
	"web-analyzer/internal/model"
	"web-analyzer/internal/repository"

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
	doc.HTMLVersion = getHTMLVersion(resp.Body)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	doc.Body = string(bodyBytes)
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
