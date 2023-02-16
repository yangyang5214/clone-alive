package types

type Options struct {
	Url         string
	MaxDepth    int8
	Headless    bool
	Append      bool
	Debug       bool
	MaxDuration int
	Concurrent  int
	Proxy       string
	Timeout     int
}

type AliveOption struct {
	Port      int
	HomeDir   string
	RouteFile string
	Debug     bool
}
