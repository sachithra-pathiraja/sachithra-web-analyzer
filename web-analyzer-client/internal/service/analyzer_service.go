package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
		return nil, err
	}

	resp, err := http.Post(
		s.analyzerURL,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// ❗ If API returned an error
	if resp.StatusCode != http.StatusOK {

		var apiErr model.APIError

		if err := json.Unmarshal(body, &apiErr); err != nil {
			return nil, fmt.Errorf("server error")
		}

		return nil, fmt.Errorf(apiErr.Message)
	}

	// ✅ Normal success response
	var result model.Response

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
