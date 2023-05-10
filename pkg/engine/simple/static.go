package simple

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	urlutil "github.com/yangyang5214/gou/url"

	rod_util "github.com/go-rod/rod/lib/utils"
	"github.com/projectdiscovery/gologger"
	"github.com/yangyang5214/clone-alive/pkg/parser"
	"github.com/yangyang5214/clone-alive/pkg/types"
	"github.com/yangyang5214/clone-alive/pkg/utils"
	"github.com/yangyang5214/gou/stack"
)

type Crawler struct {
	httpClient   *http.Client
	pendingQueue stack.Stack[string]
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
		pendingQueue: *stack.NewStack[string](),
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   60 * time.Second},
	}, nil
}

func (c *Crawler) Crawl(rootUrl string) error {
	c.pendingQueue.Push(rootUrl)
	callback := c.navigateCallback()
	for {
		if c.pendingQueue.Len() == 0 {
			gologger.Info().Msg("Url pending queue is empty, break")
			break
		}
		_url, ok := c.pendingQueue.Pop()
		if !ok {
			continue
		}
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
	if resp.BodyBytes == nil || len(resp.BodyBytes) == 0 {
		return nil
	}
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
		if urlutil.IsStaticFile(utils.GetUrlPath(req.Url)) {
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
