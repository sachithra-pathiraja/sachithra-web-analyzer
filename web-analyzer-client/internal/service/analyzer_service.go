package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"web-analyzer-client/internal/apierror"
	"web-analyzer-client/internal/model"
)

type AnalyzerService struct {
	analyzerURL string
}

func NewAnalyzerService(url string) *AnalyzerService {
	return &AnalyzerService{
		analyzerURL: url,
	}
}

func (s *AnalyzerService) CallAnalyzer(url string) (*model.Response, error) {

	reqBody := model.Request{
		URL: url,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, apierror.New(apierror.ErrRequestCreation, fmt.Sprintf("failed marshalling request: %v", err), url)
	}

	resp, err := http.Post(
		s.analyzerURL,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, apierror.New(apierror.ErrRequestFailed, fmt.Sprintf("request failed: %v", err), url)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apierror.New(apierror.ErrReadFailed, fmt.Sprintf("failed reading response: %v", err), url)
	}

	// ❗ If API returned an error
	if resp.StatusCode != http.StatusOK {

		var apiErr model.APIError

		if err := json.Unmarshal(body, &apiErr); err != nil {
			return nil, apierror.New(apierror.ErrInvalidResponse, "server returned error and response could not be parsed", url)
		}

		// map server-side error into client-side apierror
		return nil, apierror.New(apierror.ErrServerError, apiErr.Message, url)
	}

	// ✅ Normal success response
	var result model.Response

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, apierror.New(apierror.ErrUnmarshalFailed, fmt.Sprintf("failed parsing response: %v", err), url)
	}

	return &result, nil
}
