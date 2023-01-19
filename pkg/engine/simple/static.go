package simple

import (
	"clone-alive/pkg/types"
)

//todo 静态爬虫

type Crawler struct {
	option *types.Options
}

//New is created new Crawler
func New(options *types.Options) (*Crawler, error) {
	return &Crawler{
		option: options,
	}, nil
}

//Crawl crawls a URL with the specified options
func (c *Crawler) Crawl(rootURL string) error {
	//todo
	return nil
}

// Close closes the crawler process
func (c *Crawler) Close() error {
	return nil
}
