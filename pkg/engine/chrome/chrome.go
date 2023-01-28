package chrome

import (
	"context"
	"encoding/base64"
	"fmt"
	stack "github.com/emirpasic/gods/stacks/linkedliststack"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/go-rod/rod/lib/proto"
	rod_util "github.com/go-rod/rod/lib/utils"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/gologger"
	"github.com/remeh/sizedwaitgroup"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/yangyang5214/clone-alive/pkg/magic"
	"github.com/yangyang5214/clone-alive/pkg/output"
	"github.com/yangyang5214/clone-alive/pkg/parser"
	"github.com/yangyang5214/clone-alive/pkg/types"
	"github.com/yangyang5214/clone-alive/pkg/utils"
	"go.uber.org/multierr"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Crawler struct {
	browser      *rod.Browser
	tempDir      string
	targetDir    string
	rootHost     string
	domain       string
	pendingQueue stack.Stack
	urlMap       sync.Map
	mu           sync.Mutex
	option       types.Options
	expandClient *magic.ExpandVerifyCode
	outputWriter output.Writer
	previousPIDs map[int32]struct {
	} // track already running PIDs
}

// New is created new Crawler
func New(options *types.Options) (*Crawler, error) {
	if options.Url == "" {
		return nil, errors.Errorf("Url missing")
	}
	dataStore, _ := os.MkdirTemp("", "clone-alive-*")
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

	if _, err := os.Stat(targetDir); err == nil {
		_ = os.RemoveAll(targetDir)
	}

	err = os.MkdirAll(targetDir, os.ModePerm)
	if err != nil {
		return nil, err
	}

	chromeLauncher := launcher.New().
		Leakless(false).
		Set("disable-gpu", "true").
		Set("ignore-certificate-errors", "true").
		Set("ignore-certificate-errors", "1").
		Set("disable-crash-reporter", "true").
		Set("disable-notifications", "true").
		Set("hide-scrollbars", "true").
		Set("window-size", fmt.Sprintf("%d,%d", 1080, 1920)).
		Set("mute-audio", "true").
		Delete("use-mock-keychain").
		UserDataDir(dataStore)

	if options.Proxy != "" {
		chromeLauncher.Set(flags.ProxyServer, options.Proxy)
	}

	chromeLauncher.Set(flags.NoSandbox, "true")
	if options.Debug {
		chromeLauncher.Headless(false)
	}

	launcherURL, err := chromeLauncher.Launch()
	if err != nil {
		return nil, err
	}

	browser := rod.New().ControlURL(launcherURL)
	if browserErr := browser.Connect(); browserErr != nil {
		return nil, browserErr
	}

	outputWriter, err := output.New(targetDir)
	if err != nil {
		return nil, err
	}

	previousPIDs := findChromeProcesses()
	return &Crawler{
		option:       *options,
		browser:      browser,
		previousPIDs: previousPIDs,
		tempDir:      dataStore,
		targetDir:    targetDir,
		outputWriter: outputWriter,
		expandClient: magic.NewExpand(),
		rootHost:     utils.GetUrlHost(options.Url),
		domain:       utils.GetDomain(options.Url),
		pendingQueue: *stack.New(),
		urlMap:       sync.Map{},
		mu:           sync.Mutex{},
	}, nil
}

func (c *Crawler) isCrawled(urlStr string) bool {
	urls := []string{
		urlStr, urlStr + "/", strings.TrimRight(urlStr, "/"),
	}
	for _, item := range urls {
		_, exist := c.urlMap.Load(item)
		if exist {
			return true
		}
	}
	return false
}

func (c *Crawler) AddNewUrl(request types.Request) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isCrawled(request.Url) {
		return false
	}
	c.pendingQueue.Push(request)
	c.urlMap.Store(request.Url, true)
	return true
}

// Crawl crawls a URL with the specified options
func (c *Crawler) Crawl(rootURL string) error {
	ctx, cancel := context.WithCancel(context.Background())
	ctx, cancel = context.WithTimeout(ctx, time.Duration(c.option.MaxDuration)*time.Second)
	defer cancel()

	browserInstance, err := c.browser.Incognito()
	if err != nil {
		panic(err)
	}

	wg := sizedwaitgroup.New(c.option.Concurrent)
	running := int32(0)

	c.AddNewUrl(types.Request{
		Url:   rootURL,
		Depth: 0,
	})
	callback := c.navigateCallback()

	for {
		if !(atomic.LoadInt32(&running) > 0) && (c.pendingQueue.Size() == 0) {
			gologger.Info().Msg("Url pending queue is empty, break")
			break
		}

		item, ok := c.pendingQueue.Pop()
		if !ok {
			continue
		}

		req := item.(types.Request)

		wg.Add()
		atomic.AddInt32(&running, 1)

		go func() {
			defer wg.Done()
			defer atomic.AddInt32(&running, -1)

			resp, err := c.navigateRequest(browserInstance, req)
			if err != nil {
				errResult := types.ResponseResult{
					Timestamp: time.Now(),
					Url:       req.Url,
					Error:     err.Error(),
				}
				_ = c.outputWriter.Write(errResult)
				return
			}
			parser.ParseResponse(*resp, callback)
		}()
	}

	wg.Wait()

	return nil
}

// navigateCallback is add new url to queue
func (c *Crawler) navigateCallback() func(req types.Request) {
	return func(req types.Request) {
		if !utils.IsURL(req.Url) {
			resultUrl, err := url.JoinPath(c.domain, req.Url)
			if err != nil {
				return
			}
			req.Url = resultUrl
		}
		if strings.Contains(req.Url, "javascript:") {
			return
		}
		urlParsed, err := url.Parse(req.Url)
		if err != nil {
			return
		}
		if urlParsed.Host != c.rootHost {
			return
		}
		if c.AddNewUrl(req) {
			gologger.Info().Msgf("find new url %s, depth %d", req.Url, req.Depth)
		}
	}
}

// navigateRequest is process single url
func (c *Crawler) navigateRequest(browser *rod.Browser, req types.Request) (*types.Response, error) {
	page, err := browser.Page(proto.TargetCreateTarget{URL: req.Url})
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	defer page.Close()

	lastTimestamp := time.Now().Unix()

	var resp *types.ResponseResult //req.url response

	requestMap := sync.Map{}

	go page.EachEvent(func(e *proto.NetworkLoadingFinished) {
		data, ok := requestMap.Load(e.RequestID)
		if !ok {
			return
		}
		event := data.(*types.EventListen)
		request := event.Request
		response := event.Response

		if request == nil {
			request = &proto.NetworkRequest{
				Method: http.MethodGet,
				URL:    response.URL,
			}
		}
		_url := request.URL
		if c.isCrawled(_url) {
			return
		}
		urlParsed, err := url.Parse(_url)
		if err != nil {
			return
		}

		if urlParsed.Host != c.rootHost {
			return // 外部站点
		}

		m := proto.NetworkGetResponseBody{RequestID: e.RequestID}
		r, err := m.Call(page)
		if err != nil {
			return
		}
		contentType := response.MIMEType
		urlPath := urlParsed.Path
		gologger.Info().Msgf("【%s】 %s find --> %s", contentType, _url, urlPath)

		requestContentTypeVal := request.Headers["Content-Type"].Val()
		if requestContentTypeVal == nil {
			requestContentTypeVal = ""
		}

		respResult := &types.ResponseResult{
			Timestamp:           time.Now(),
			Url:                 _url,
			Body:                r.Body,
			Status:              response.Status,
			HttpMethod:          request.Method,
			RequestContentType:  requestContentTypeVal.(string),
			ResponseContentType: response.MIMEType,
			Depth:               req.Depth + 1,
		}

		var expandResult = c.expandClient.Run(_url, contentType)
		if expandResult == nil {
			c.log(respResult)
		} else {
			for _, item := range expandResult {
				itemResp := *respResult
				itemResp.Body = item.Body
				itemResp.Url = item.UrlStr
				itemResp.ResponseContentType = ""
				c.saveFile(utils.GetUrlPath(itemResp.Url), &itemResp)
			}
			_ = c.outputWriter.Write(*respResult)
		}

		if utils.IsSameURL(_url, req.Url) {
			resp = respResult
		}

		lastTimestamp = time.Now().Unix()
	}, func(e *proto.NetworkRequestWillBeSent) {
		requestMap.Store(e.RequestID, &types.EventListen{
			Request: e.Request,
		})
	}, func(e *proto.NetworkResponseReceived) {
		data, ok := requestMap.Load(e.RequestID)
		if !ok {
			requestMap.Store(e.RequestID, &types.EventListen{
				Response: e.Response,
			})
		} else {
			event := data.(*types.EventListen)
			event.Response = e.Response
		}
	})()

	page.MustWaitNavigation()
	page.Timeout(time.Duration(c.option.Timeout) * time.Second)
	page.MustWaitLoad()
	page.MustWaitIdle()

	page.MustReload() //reload page

	for {
		if time.Now().Unix()-lastTimestamp > 5 {
			break
		}
		time.Sleep(time.Millisecond * 300)
	}

	if resp == nil {
		html, err := page.HTML()
		if err != nil {
			return nil, errors.Wrap(err, "could not get html")
		}
		resp = &types.ResponseResult{
			Timestamp:           time.Now(),
			Url:                 req.Url, //currentUrl already collect in network event
			Body:                html,
			Status:              http.StatusOK,
			ResponseContentType: types.TextHtml,
			HttpMethod:          http.MethodGet,
			Depth:               req.Depth + 1,
		}
	}

	if resp.ResponseContentType == types.TextHtml && utils.GetUrlPath(req.Url) == utils.GetUrlPath(c.option.Url) {
		page.MustScreenshotFullPage(filepath.Join(c.targetDir, "screenshot", utils.GetUrlHost(req.Url)+".png"))
	}
	c.log(resp)
	return &types.Response{
		Body:  resp.Body,
		Depth: req.Depth + 1,
	}, nil
}

func (c *Crawler) log(result *types.ResponseResult) {
	c.urlMap.Store(result.Url, true)
	_ = c.outputWriter.Write(*result)
	c.saveFile(utils.GetUrlPath(result.Url), result)
	if utils.IsSameURL(result.Url, c.option.Url) {
		parsed, _ := url.Parse(result.Url)
		result.Url = fmt.Sprintf(`%s://%s`, parsed.Scheme, parsed.Host)
		_ = c.outputWriter.Write(*result)
		c.saveFile("index.html", result)
	}
}

func (c *Crawler) locationHref(page *rod.Page) (string, error) {
	res, err := page.Eval(`() => location.href`)
	if err != nil {
		return "", err
	}
	return res.Value.String(), nil
}

// getScrollHeight it is get 'document.body.scrollHeight'
func (c *Crawler) getScrollHeight(page *rod.Page) int {
	res, _ := page.Eval(`document.body.scrollHeight`)
	return res.Value.Int()
}

// saveFile it's save data to file
func (c *Crawler) saveFile(urlPath string, resp *types.ResponseResult) {
	var data interface{}
	data = resp.Body

	paths := []string{c.targetDir, urlPath}

	//https://github.com/yangyang5214/clone-alive/issues/15
	if resp.ResponseContentType == "text/html" && urlPath != "index.html" {
		paths = append(paths, "index.html")
	}

	//replace original url
	resp.Body = strings.Replace(resp.Body, c.domain, "", -1)
	resp.Body = strings.Replace(resp.Body, c.rootHost, "", -1)
	resp.Body = strings.Replace(resp.Body, strings.Split(c.rootHost, ":")[0], "", -1)

	if strings.HasPrefix(resp.ResponseContentType, "image") {
		data = base64.NewDecoder(base64.StdEncoding, strings.NewReader(data.(string)))
	}
	err := rod_util.OutputFile(filepath.Join(paths...), data)
	if err != nil {
		gologger.Error().Msgf("OutputFile error: %s", err.Error())
	}
}

// Close closes the crawler process
func (c *Crawler) Close() error {
	if err := c.browser.Close(); err != nil {
		return err
	}

	if err := os.RemoveAll(c.tempDir); err != nil {
		return err
	}

	_ = c.outputWriter.Close()

	return c.killChromeProcesses()
}

// killChromeProcesses any and all new chrome processes started after
// headless process launch.
func (c *Crawler) killChromeProcesses() error {
	var errs []error
	processes, _ := process.Processes()

	for _, p := range processes {
		// skip non-chrome processes
		if !isChromeProcess(p) {
			continue
		}

		// skip chrome processes that were already running
		if _, ok := c.previousPIDs[p.Pid]; ok {
			continue
		}

		if err := p.Kill(); err != nil {
			gologger.Info().Msgf("kill chrome process error %d, %s", p.Pid, err.Error())
			errs = append(errs, err)
		}
	}

	return multierr.Combine(errs...)
}

// findChromeProcesses finds chrome process running on host
func findChromeProcesses() map[int32]struct{} {
	processes, _ := process.Processes()
	list := make(map[int32]struct{})
	for _, p := range processes {
		if isChromeProcess(p) {
			list[p.Pid] = struct{}{}
			if ppid, err := p.Ppid(); err == nil {
				list[ppid] = struct{}{}
			}
		}
	}
	return list
}

// isChromeProcess checks if a process is chrome/chromium
func isChromeProcess(process *process.Process) bool {
	name, _ := process.Name()
	if name == "" {
		return false
	}
	return strings.HasPrefix(strings.ToLower(name), "chromium")
}
