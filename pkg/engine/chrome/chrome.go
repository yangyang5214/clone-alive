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
	"github.com/yangyang5214/clone-alive/pkg/output"
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
	option       types.Options
	outputWriter output.Writer
	previousPIDs map[int32]struct{} // track already running PIDs
}

//New is created new Crawler
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
	}, nil
}

//Crawl crawls a URL with the specified options
func (c *Crawler) Crawl(rootURL string) error {
	ctx, cancel := context.WithCancel(context.Background())
	ctx, cancel = context.WithTimeout(ctx, time.Duration(c.option.MaxDuration)*time.Second)
	defer cancel()

	urlParsed, _ := url.Parse(rootURL)
	browserInstance, err := c.browser.Incognito()
	if err != nil {
		panic(err)
	}

	wg := sizedwaitgroup.New(c.option.Concurrent)
	running := int32(0)

	pendingQueue := stack.New()

	pendingQueue.Push(types.Request{
		Url:       rootURL,
		UrlParsed: urlParsed,
	})
	callback := c.navigateCallback(pendingQueue)

	for {
		if !(atomic.LoadInt32(&running) > 0) && (pendingQueue.Size() == 0) {
			gologger.Info().Msg("Url pending queue is empty, break")
			break
		}

		item, ok := pendingQueue.Pop()
		if !ok {
			continue
		}

		req := item.(types.Request)

		wg.Add()
		atomic.AddInt32(&running, 1)

		go func() {
			defer wg.Done()
			defer atomic.AddInt32(&running, -1)

			err := c.navigateRequest(browserInstance, req, callback, urlParsed.Host)
			if err != nil {
				errResult := types.ResponseResult{
					Timestamp: time.Now(),
					Url:       req.Url,
					Error:     err.Error(),
				}
				_ = c.outputWriter.Write(errResult)
			}
		}()
	}

	wg.Wait()

	return nil
}

//navigateRequest is process single url
func (c *Crawler) navigateRequest(browser *rod.Browser, req types.Request, callback func(r types.Request), rootHost string) error {
	page := browser.MustPage(req.Url)

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
		urlParsed, err := url.Parse(_url)
		if err != nil {
			return
		}

		if urlParsed.Host != rootHost {
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

		respResult := &types.ResponseResult{
			Timestamp:           time.Now(),
			Url:                 _url,
			Body:                r.Body,
			Status:              response.Status,
			HttpMethod:          request.Method,
			RequestContentType:  request.Headers["Content-Type"].String(),
			ResponseContentType: response.MIMEType,
		}

		c.log(respResult)
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
	page.MustWaitLoad()
	page.MustWaitIdle()

	page.MustReload() //reload page

	for {
		if time.Now().Unix()-lastTimestamp > 3 {
			break
		}
		time.Sleep(1)
	}

	if resp == nil {
		html, err := page.HTML()
		if err != nil {
			return errors.Wrap(err, "could not get html")
		}
		resp = &types.ResponseResult{
			Timestamp:           time.Now(),
			Url:                 req.Url, //currentUrl already collect in network event
			Body:                html,
			Status:              http.StatusOK,
			ResponseContentType: types.TextHtml,
			HttpMethod:          http.MethodGet,
		}
	}

	if resp.ResponseContentType == types.TextHtml {
		page.MustScreenshotFullPage(filepath.Join(c.targetDir, "screenshot", req.UrlParsed.Host+".png"))
	}
	c.log(resp)
	return nil
}

func (c *Crawler) log(result *types.ResponseResult) {
	_ = c.outputWriter.Write(*result)
	c.saveFile(utils.GetUrlPath(result.Url), result)
	if utils.IsSameURL(result.Url, c.option.Url) {
		parsed, _ := url.Parse(result.Url)
		result.Url = fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
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

//getScrollHeight it is get 'document.body.scrollHeight'
func (c *Crawler) getScrollHeight(page *rod.Page) int {
	res, _ := page.Eval(`document.body.scrollHeight`)
	return res.Value.Int()
}

//saveFile it's save data to file
func (c *Crawler) saveFile(urlPath string, resp *types.ResponseResult) {
	var data interface{}
	data = resp.Body
	if urlPath == "" || urlPath == "/" {
		urlPath = "index.html"
	}
	paths := []string{c.targetDir, urlPath}

	if strings.HasPrefix(resp.ResponseContentType, "image") {
		data = base64.NewDecoder(base64.StdEncoding, strings.NewReader(data.(string)))
	}
	_ = rod_util.OutputFile(filepath.Join(paths...), data)
}

//navigateCallback is add new url to queue //todo
func (c *Crawler) navigateCallback(queue *stack.Stack) func(r types.Request) {
	return func(r types.Request) {

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
