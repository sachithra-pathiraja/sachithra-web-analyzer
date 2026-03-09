package service

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"web-analyzer/internal/apierror"
	"web-analyzer/internal/model"

	"github.com/PuerkitoBio/goquery"
)

type FetchService struct {
	client      *http.Client
	logger      *slog.Logger
	linkWorkers int
}

func NewFetchService(client *http.Client, logger *slog.Logger, workers int) *FetchService {
	return &FetchService{
		client:      client,
		logger:      logger,
		linkWorkers: workers,
	}
}

func (s *FetchService) ProcessDocument(ctx context.Context, doc *model.Document) (*model.Document, error) {
	s.logger.Info("Processing document", "url", doc.URL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, doc.URL, nil)
	if err != nil {
		s.logger.Error("invalid url", "url", doc.URL, "error", err)
		return nil, apierror.New(
			apierror.ErrInvalidURL,
			"invalid url provided",
		)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("failed fetching document", "url", doc.URL, "error", err)
		return nil, apierror.New(
			apierror.ErrFetchFailed,
			"failed to fetch document",
		)
	}
	defer resp.Body.Close()
	// Read the response body once and reuse readers for parsing
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("failed reading response body", "error", err)

		return nil, apierror.New(
			apierror.ErrInternal,
			"failed reading response body",
		)
	}

	rdr := bytes.NewReader(bodyBytes)
	doc.HTMLVersion, err = getHTMLVersion(rdr, s.logger)
	if err != nil {
		s.logger.Error("failed getting HTML version", "error", err)

		return nil, apierror.New(
			apierror.ErrInternal,
			"failed getting HTML version",
		)
	}

	// reuse the reader for title extraction
	if _, err := rdr.Seek(0, io.SeekStart); err != nil {
		// if seek fails, try creating a new reader
		rdr = bytes.NewReader(bodyBytes)
	}
	docFromReader, err := goquery.NewDocumentFromReader(rdr)
	if err != nil {
		s.logger.Error("failed parsing html", "error", err)

		return nil, apierror.New(
			apierror.ErrParseFailed,
			"failed parsing html document",
		)
	}
	doc.Title, doc.Headings, err = getTitleAndHeadings(docFromReader, s.logger)
	if err != nil {
		s.logger.Error("failed getting title and headings", "error", err)

		return nil, apierror.New(
			apierror.ErrInternal,
			"failed getting title and headings",
		)
	}
	doc.Links, err = getLinks(docFromReader, doc.URL, s.client, s.logger, s.linkWorkers)
	if err != nil {
		s.logger.Error("failed getting links", "error", err)

		return nil, apierror.New(
			apierror.ErrInternal,
			"failed getting links",
		)
	}
	doc.HasLoginForm = getHasLogin(docFromReader)

	return doc, nil
}
