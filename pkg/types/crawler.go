package types

import "github.com/go-rod/rod/lib/proto"

type CrawlerType = string

const (
	Chrome CrawlerType = "chrome"
	Simple CrawlerType = "simple"
)

type EventListen struct {
	Request   *proto.NetworkRequest
	Response  *proto.NetworkResponse
	RequestId proto.NetworkRequestID
}
