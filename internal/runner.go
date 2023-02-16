package internal

import (
	"github.com/pkg/errors"
	"github.com/projectdiscovery/gologger"
	"github.com/yangyang5214/clone-alive/internal/banner"
	"github.com/yangyang5214/clone-alive/pkg/engine"
	"github.com/yangyang5214/clone-alive/pkg/engine/chrome"
	"github.com/yangyang5214/clone-alive/pkg/engine/simple"
	"github.com/yangyang5214/clone-alive/pkg/types"
	"github.com/yangyang5214/clone-alive/pkg/utils"
	"net/url"
	"os"
	"path"
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

	urlParsed, err := url.Parse(options.Url)
	if err != nil {
		return nil, err
	}

	if options.Proxy != "" {
		_, err = url.Parse(options.Proxy)
		if err != nil {
			return nil, errors.Wrap(err, "proxy url error")
		}
	}

	targetDir := path.Join(utils.CurrentDirectory(), urlParsed.Host)

	if !options.Append {
		if _, err := os.Stat(targetDir); err == nil {
			_ = os.RemoveAll(targetDir)
		}

		err = os.MkdirAll(targetDir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	options.TargetDir = targetDir

	crawler, err = chrome.New(options)
	if err != nil {
		return nil, errors.Wrap(err, "could not create crawler engine")
	}

	if options.Static {
		crawler, err = simple.New(options)
	} else {
		crawler, err = chrome.New(options)
	}

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
