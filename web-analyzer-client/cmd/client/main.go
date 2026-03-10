package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"web-analyzer-client/internal/config"
	"web-analyzer-client/internal/handler"
	"web-analyzer-client/internal/service"
)

func main() {

	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	analyzerService := service.NewAnalyzerService(cfg.Analyzer.URL)

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("/", handler.HomeHandler(tmpl))
	mux.HandleFunc("/analyze", handler.AnalyzeHandler(tmpl, analyzerService))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  300 * time.Second,
		WriteTimeout: 300 * time.Second,
	}

	log.Printf("Client running on http://localhost:%d", cfg.Server.Port)

	log.Fatal(server.ListenAndServe())
}
