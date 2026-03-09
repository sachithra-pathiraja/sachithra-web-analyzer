package service

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
	"web-analyzer/internal/apierror"
	"web-analyzer/internal/model"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func getHTMLVersion(r io.Reader, logger *slog.Logger) (string, error) {

	tokenizer := html.NewTokenizer(r)

	for {
		tt := tokenizer.Next()

		if tt == html.DoctypeToken {

			token := tokenizer.Token()
			raw := token.String()

			logger.Info("doctype detected", "raw", raw)

			if token.Data == "html" && len(token.Attr) == 0 {
				return "HTML5", nil
			}

			if strings.Contains(raw, "XHTML 1.0") {
				return "XHTML 1.0", nil
			}

			if strings.Contains(raw, "HTML 4.01") {
				return "HTML 4.01", nil
			}

			return "Legacy / Unknown DOCTYPE", nil
		}

		if tt == html.ErrorToken {

			err := tokenizer.Err()

			if err != io.EOF {
				logger.Error("doctype parsing error", "error", err)

				return "", apierror.New(
					apierror.ErrParseFailed,
					"failed parsing html doctype",
				)
			}

			break
		}
	}

	logger.Info("doctype not found")

	return "No DOCTYPE (Quirks Mode)", nil
}

func getTitleAndHeadings(doc *goquery.Document, logger *slog.Logger) (string, []model.Heading, error) {

	var headings []model.Heading

	for i := 1; i <= 6; i++ {

		tag := fmt.Sprintf("h%d", i)

		count := doc.Find(tag).Length()

		if count > 0 {
			headings = append(headings, model.Heading{
				Level: i,
				Count: count,
			})
		}
	}

	title := doc.Find("title").Text()

	logger.Info("document headings analyzed",
		"title", title,
		"heading_count", len(headings),
	)

	return title, headings, nil
}
