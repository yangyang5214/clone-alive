package types

type Options struct {
	Url         string
	MaxDepth    int8
	Static      bool
	Append      bool
	Debug       bool
	MaxDuration int
	Concurrent  int
	Proxy       string
	Timeout     int
	TargetDir   string
}

type AliveOption struct {
	Port      int
	HomeDir   string
	RouteFile string
	Debug     bool
}
