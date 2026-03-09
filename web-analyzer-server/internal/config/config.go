package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerPort  string
	AnalyzerURL string
	LinkWorkers int
}

func Load(path string) (*Config, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := map[string]string{}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)

		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			values[key] = val
		}
	}

	workers, _ := strconv.Atoi(values["LINK_WORKERS"])
	cfg := &Config{
		ServerPort:  values["SERVER_PORT"],
		AnalyzerURL: values["ANALYZER_URL"],
		LinkWorkers: workers,
	}

	return cfg, nil
}
