package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"
	"time"
)

type Request struct {
	URL string `json:"URL"`
}

const page = `
<!DOCTYPE html>
<html>
<head>
	<title>Web Analyzer Client</title>
</head>
<body>
	<h2>Web Analyzer</h2>
	<form method="POST" action="/analyze">
		<input type="text" name="url" placeholder="Paste URL here" size="50" required>
		<button type="submit">Analyze</button>
	</form>

	{{if .}}
	<h3>Response:</h3>
	<pre>{{.Beautified}}</pre>
	{{end}}
</body>
</html>
`

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/analyze", analyzeHandler)

	server := &http.Server{
		Addr:         ":8090",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// ---- Start server in goroutine ----
	go func() {
		log.Println("Client UI running on http://localhost:8090")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	// ---- Listen for shutdown signal ----
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutdown signal received")

	// ---- Graceful shutdown ----
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Graceful shutdown failed: %v", err)
	}

	log.Println("Server stopped gracefully")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("page").Parse(page))
	tmpl.Execute(w, nil)
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {

	url := r.FormValue("url")

	reqBody := Request{
		URL: url,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.Post(
		"http://localhost:8080/analyzer",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var respObj interface{}
	var beautified string

	if err := json.Unmarshal(body, &respObj); err == nil {

		pretty, err := json.MarshalIndent(respObj, "", "  ")
		if err == nil {
			beautified = string(pretty)
		} else {
			beautified = string(body)
		}

	} else {
		beautified = string(body)
	}

	tmpl := template.Must(template.New("page").Parse(page))
	tmpl.Execute(w, map[string]interface{}{"Beautified": beautified})
}
