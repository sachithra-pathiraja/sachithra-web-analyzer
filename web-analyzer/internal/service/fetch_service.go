package service

import (
	"io"
	"log"
	"net/http"
	"web-analyzer/internal/model"
	"web-analyzer/internal/repository"
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
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	doc.Body = string(bodyBytes)
	return doc, nil
}
