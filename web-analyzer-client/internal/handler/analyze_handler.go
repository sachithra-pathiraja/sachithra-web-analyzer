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

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, result)
	}
}
