package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"text/template"
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
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/analyze", analyzeHandler)

	log.Println("Client UI running on http://localhost:8090")
	log.Fatal(http.ListenAndServe(":8090", nil))
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
