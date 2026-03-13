package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"web-analyzer/internal/apierror"
	"web-analyzer/internal/model"
	"web-analyzer/internal/service"
)

type AnalyzerHandler struct {
	processor service.DocumentProcessor
}

func NewAnalyzerHandler(p service.DocumentProcessor) *AnalyzerHandler {
	return &AnalyzerHandler{processor: p}
}

func (h *AnalyzerHandler) Analyze(w http.ResponseWriter, r *http.Request) {

	var doc model.Document

	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		writeError(w, http.StatusBadRequest,
			apierror.ErrInvalidURL,
			"Invalid request body")
		return
	}

	result, err := h.processor.ProcessDocument(r.Context(), &doc)
	if err != nil {

		var apiErr *apierror.Error

		if errors.As(err, &apiErr) {

			status := mapErrorToHTTP(apiErr.Code)

			writeError(w, status, apiErr.Code, apiErr.Message)
			return
		}

		writeError(w, http.StatusInternalServerError,
			apierror.ErrInternal,
			"Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {

	resp := map[string]string{
		"code":    code,
		"message": message,
	}

	writeJSON(w, status, resp)
}

func mapErrorToHTTP(code string) int {

	switch code {

	case apierror.ErrInvalidURL:
		return http.StatusBadRequest

	// Upstream / remote failures
	case apierror.ErrFetchFailed,
		apierror.ErrReadFailed,
		apierror.ErrRequestFailed,
		apierror.ErrRequestCreation,
		apierror.ErrInaccessibleLink,
		apierror.ErrLinkAnalysisFailed:
		return http.StatusBadGateway

	// Timeouts / cancellations
	case apierror.ErrRequestTimeout:
		return http.StatusGatewayTimeout

	// Parse / extraction issues
	case apierror.ErrHTMLParseFailed,
		apierror.ErrExtractionFailed,
		apierror.ErrParseFailed:
		return http.StatusUnprocessableEntity

	default:
		return http.StatusInternalServerError
	}
}
