package main

import (
	"net/http"
	"web-analyzer/internal/database"
	"web-analyzer/internal/handler"
	"web-analyzer/internal/repository"
	"web-analyzer/internal/service"
)

func main() {
	db := database.NewMySQL()

	repo := repository.NewDocumentRepository(db)
	fetchService := service.NewFetchService(repo)
	h := handler.NewDocumentHandler(fetchService)

	http.HandleFunc("/analyzer", h.Handle)
	http.ListenAndServe(":8080", nil)
}
