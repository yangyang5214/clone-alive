package types

type Options struct {
	Url         string
	MaxDepth    int8
	Headless    bool
	Debug       bool
	MaxDuration int
	Concurrent  int
	Proxy       string
}
