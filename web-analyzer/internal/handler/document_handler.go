package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"web-analyzer/internal/model"
	"web-analyzer/internal/service"
)

type DocumentHandler struct {
	service *service.FetchService
}

func NewDocumentHandler(s *service.FetchService) *DocumentHandler {
	return &DocumentHandler{service: s}
}

func (h *DocumentHandler) Handle(w http.ResponseWriter, r *http.Request) {
	log.Println("Received /collect request")
	var document model.Document

	if err := json.NewDecoder(r.Body).Decode(&document); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	log.Printf("Decoded document: %+v", document)
	processedDoc, err := h.service.ProcessDocument(&document)
	if err != nil {
		log.Printf("Error processing document: %v", err)
		http.Error(w, "processing failed", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(processedDoc)
}
