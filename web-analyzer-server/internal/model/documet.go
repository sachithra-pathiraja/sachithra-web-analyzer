package model

type Document struct {
	URL          string
	HTMLVersion  string
	Title        string
	Headings     []Heading
	Links        []Link
	HasLoginForm bool
}
