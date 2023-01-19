package types

import "net/url"

type Request struct {
	Url       string
	UrlParsed *url.URL
}

type Response struct {
	Url string
}
