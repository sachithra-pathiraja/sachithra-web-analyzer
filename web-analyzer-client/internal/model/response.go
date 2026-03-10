package model

type Request struct {
	URL string `json:"URL"`
}

type Response struct {
	HTMLVersion  string    `json:"htmlVersion"`
	Title        string    `json:"title"`
	Headings     []Heading `json:"headings"`
	Links        []Link    `json:"links"`
	HasLoginForm bool      `json:"hasLoginForm"`
}

type Heading struct {
	Level int `json:"level"`
	Count int `json:"count"`
}

type Link struct {
	LinkType string `json:"linkType"`
	Count    int    `json:"count"`
}
