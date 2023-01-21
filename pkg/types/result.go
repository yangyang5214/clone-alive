package types

import (
	"time"
)

type RequestResult struct {
	Timestamp time.Time `json:"timestamp"`
	Url       string    `json:"url,omitempty"`
	Method    string
}

type ResponseResult struct {
	Timestamp           time.Time `json:"timestamp"`
	HttpMethod          string    `json:"http_method"`
	Url                 string    `json:"url,omitempty"`
	BodyLen             int       `json:"body_len"`
	Title               string    `json:"title"`
	RequestContentType  string    `json:"request_content_type"`
	ResponseContentType string    `json:"response_content_type"`
	Body                string    `json:"body"`
	Status              int       `json:"status"`
	Error               string    `json:"error,omitempty"`
}
