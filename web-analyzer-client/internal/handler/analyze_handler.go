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
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		url := r.FormValue("url")

		result, err := svc.CallAnalyzer(url)

		data := map[string]interface{}{
			"URL": url,
		}

		if err != nil {
			data["Error"] = err
		} else {
			data["Result"] = result
		}

		tmpl.Execute(w, data)
	}
}
