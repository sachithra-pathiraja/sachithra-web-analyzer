package service

import (
	"bytes"
	"encoding/json"
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

	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post(
		s.analyzerURL,
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result model.Response

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
