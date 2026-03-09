package service

import (
	"context"
	"web-analyzer/internal/model"
)

type DocumentProcessor interface {
	ProcessDocument(ctx context.Context, doc *model.Document) (*model.Document, error)
}
