package types

import (
	"github.com/PuerkitoBio/goquery"
)

type Request struct {
	Url   string
	Depth int
}

type Response struct {
	Body   string
	Reader *goquery.Document
	Depth  int
}
