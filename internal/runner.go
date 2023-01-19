package internal

import (
	"github.com/pkg/errors"
	"github.com/projectdiscovery/gologger"
	"github.com/yangyang5214/clone-alive/internal/banner"
	"github.com/yangyang5214/clone-alive/pkg/engine"
	"github.com/yangyang5214/clone-alive/pkg/engine/chrome"
	"github.com/yangyang5214/clone-alive/pkg/engine/simple"
	"github.com/yangyang5214/clone-alive/pkg/types"
)

type Runner struct {
	option  *types.Options
	crawler engine.Engine
}

func New(options *types.Options) (*Runner, error) {
	banner.ShowBanner()
	var (
		crawler engine.Engine
		err     error
	)

	var crawlerType types.CrawlerType
	if options.Headless {
		crawler, err = chrome.New(options)
		crawlerType = types.Chrome
	} else {
		crawler, err = simple.New(options)
		crawlerType = types.Simple
	}
	if err != nil {
		return nil, errors.Wrap(err, "could not create crawler engine")
	}

	gologger.Info().Msgf("Start crawler with <%s>", crawlerType)
	return &Runner{
		option:  options,
		crawler: crawler,
	}, nil
}

func (r *Runner) Run() {
	defer r.crawler.Close()

	if err := r.crawler.Crawl(r.option.Url); err != nil {
		gologger.Error().Msgf("Could not crawl %s: %s", r.option.Url, err)
	}
}
