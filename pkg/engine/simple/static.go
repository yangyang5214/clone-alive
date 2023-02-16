package simple

import (
	"crypto/tls"
	stack "github.com/emirpasic/gods/stacks/linkedliststack"
	rod_util "github.com/go-rod/rod/lib/utils"
	"github.com/projectdiscovery/gologger"
	"github.com/yangyang5214/clone-alive/pkg/parser"
	"github.com/yangyang5214/clone-alive/pkg/types"
	"github.com/yangyang5214/clone-alive/pkg/utils"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

type Crawler struct {
	httpClient   *http.Client
	pendingQueue stack.Stack
	domain       string
	targetDir    string
}

// New is created new Crawler
func New(option *types.Options) (*Crawler, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &Crawler{
		targetDir:    option.TargetDir,
		domain:       utils.GetDomain(option.Url),
		pendingQueue: *stack.New(),
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   60 * time.Second},
	}, nil
}

func (c *Crawler) Crawl(rootUrl string) error {
	c.pendingQueue.Push(rootUrl)
	callback := c.navigateCallback()
	for {
		if c.pendingQueue.Size() == 0 {
			gologger.Info().Msg("Url pending queue is empty, break")
			break
		}
		item, ok := c.pendingQueue.Pop()
		if !ok {
			continue
		}

		_url := item.(string)
		resp := c.CrawlAndSave(_url, utils.GetUrlPath(_url))
		if resp == nil {
			continue
		}
		parser.ParseResponse(*resp, callback)
	}
	return nil
}

func (c *Crawler) CrawlAndSave(url string, targetPath string) (resp *types.Response) {
	targetPath = filepath.Join(c.targetDir, targetPath)
	resp, err := c.navigateRequest(url)
	if err != nil {
		gologger.Error().Msgf("navigateRequest error %s", err.Error())
		return nil
	}
	if resp == nil {
		return resp
	}
	gologger.Info().Msgf("For url <%s> => <%s>", url, targetPath)
	err = rod_util.OutputFile(targetPath, resp.BodyBytes)
	if err != nil {
		return nil
	}
	return resp
}

func (c *Crawler) navigateCallback() func(req types.Request) {
	return func(req types.Request) {
		if !utils.IsURL(req.Url) {
			resultUrl, err := url.JoinPath(c.domain, req.Url)
			if err != nil {
				return
			}
			req.Url = resultUrl
		}
		c.pendingQueue.Push(req.Url)
		if types.IsStaticFile(utils.GetUrlPath(req.Url)) {
			c.pendingQueue.Push(utils.GetRealUrl(req.Url))
		}
	}
}

func (c *Crawler) navigateRequest(url string) (*types.Response, error) {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, nil
	}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &types.Response{
		Body:        string(bytes),
		BodyBytes:   bytes,
		StatusCode:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
	}, nil
}

func (c *Crawler) Close() error {
	return nil
}
