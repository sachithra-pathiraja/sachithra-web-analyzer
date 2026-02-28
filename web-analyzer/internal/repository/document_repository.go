package repository

import (
	"database/sql"
	"log"
	"web-analyzer/internal/model"
)

type DocumentRepository struct {
	db *sql.DB
}

func NewDocumentRepository(db *sql.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

func (r *DocumentRepository) Save(e *model.Document) error {
	log.Printf("Saving document to DB: %+v", e)
	query := `
	INSERT INTO document (url, body)
	VALUES (?, ?))
	`

	_, err := r.db.Exec(query,
		e.URL,
		e.Body,
	)

	if err != nil {
		log.Printf("Error executing DB insert: %v", err)
	}

	return err
}
