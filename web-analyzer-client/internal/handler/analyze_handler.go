package handler

import (
	"net/http"
	"text/template"

	"web-analyzer-client/internal/service"
)

func HomeHandler(tmpl *template.Template) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	}
}

func AnalyzeHandler(tmpl *template.Template, svc *service.AnalyzerService) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		url := r.FormValue("url")

		result, err := svc.CallAnalyzer(url)

		data := map[string]interface{}{
			"URL": url,
		}

		if err != nil {
			data["Error"] = err.Error()
		} else {
			data["Result"] = result
		}

		tmpl.Execute(w, data)
	}
}
