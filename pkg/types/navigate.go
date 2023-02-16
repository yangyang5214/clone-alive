package types

import (
	"github.com/PuerkitoBio/goquery"
)

type Request struct {
	Url   string
	Depth int
}

type Response struct {
	ContentType string
	Body        string
	StatusCode  int
	BodyBytes   []byte
	Reader      *goquery.Document
	Depth       int
}
