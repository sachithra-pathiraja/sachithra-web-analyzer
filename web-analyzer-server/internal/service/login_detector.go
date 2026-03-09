package service

import (
	"github.com/PuerkitoBio/goquery"
)

func getHasLogin(doc *goquery.Document) bool {
	return doc.Find("input[type='password']").Length() > 0
}
