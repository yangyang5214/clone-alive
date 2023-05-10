package types

type Options struct {
	Url            string
	MaxDepth       int8
	Static         bool
	Debug          bool
	MaxDuration    int
	Concurrent     int
	Proxy          string
	Timeout        int
	TargetDir      string
	VerifyCodePath string
}

type AliveOption struct {
	Port           int
	HomeDir        string
	RouteFile      string
	Debug          bool
	VerifyCodePath string
}
