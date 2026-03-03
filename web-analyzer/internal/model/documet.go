package model

type Document struct {
	URL         string
	Body        string
	HTMLVersion string
	Title       string
	Headings    []Heading
	Links       []Link
}
