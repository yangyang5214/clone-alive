package chrome

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yangyang5214/gou/set"

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
	fileutil "github.com/yangyang5214/gou/file"
	"github.com/yangyang5214/gou/stack"
	urlutil "github.com/yangyang5214/gou/url"
	"go.uber.org/multierr"
)

const (
	UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"
	Language  = "zh-CN,zh;q=0.9"
)

type Crawler struct {
	browser       *rod.Browser
	tempDir       string
	targetDir     string
	rootHost      string
	domain        string
	domains       []string
	pendingQueue  *stack.Stack[types.Request]
	crawledUrl    *set.SyncSet
	option        types.Options
	expandClient  *magic.ExpandVerifyCode
	attributeMock *magic.Attribute
	outputWriter  output.Writer
	previousPIDs  map[int32]struct {
	} // track already running PIDs
}

// New is created new Crawler
func New(options *types.Options) (*Crawler, error) {
	if options.Url == "" {
		return nil, errors.Errorf("Url missing")
	}
	dataStore, _ := os.MkdirTemp("", "clone-alive-*")

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
		Env(append(os.Environ(), "TZ=Asia/Shanghai")...).
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

	outputWriter, err := output.New(options.TargetDir)
	if err != nil {
		return nil, err
	}

	previousPIDs := findChromeProcesses()
	return &Crawler{
		option:        *options,
		browser:       browser,
		previousPIDs:  previousPIDs,
		tempDir:       dataStore,
		targetDir:     options.TargetDir,
		outputWriter:  outputWriter,
		expandClient:  magic.NewExpand(magic.RetryCount, options.VerifyCodePath),
		rootHost:      urlutil.GetUrlHost(options.Url),
		domain:        utils.GetDomain(options.Url),
		domains:       utils.GetDomains(options.Url),
		pendingQueue:  stack.NewStack[types.Request](),
		crawledUrl:    set.NewSyncSet(),
		attributeMock: magic.NewAttribute(),
	}, nil
}

func (c *Crawler) isCrawled(urlStr string) bool {
	urls := []string{
		urlStr, urlStr + "/",
		strings.TrimRight(urlStr, "/"),
		utils.GetRealUrl(urlStr),
	}
	for _, item := range urls {
		if c.crawledUrl.Contains(item) {
			return true
		}
	}
	return false
}

func (c *Crawler) AddNewUrl(request types.Request) bool {
	if c.isCrawled(request.Url) {
		return false
	}
	c.pendingQueue.Push(request)
	c.crawledUrl.Add(request.Url)
	c.crawledUrl.Add(utils.GetRealUrl(request.Url))
	return true
}

// addDefaultUrls adds default URLs to the crawler
// https://github.com/yangyang5214/clone-alive/issues/31
func (c *Crawler) addDefaultUrls() {
	defaultUrls := []string{
		"/favicon.ico",
	}
	for _, item := range defaultUrls {
		c.AddNewUrl(types.Request{
			Url:   c.domain + item,
			Depth: 0,
		})
	}
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
	c.addDefaultUrls()
	callback := c.navigateCallback()

	for {
		if !(atomic.LoadInt32(&running) > 0) && (c.pendingQueue.Len() == 0) {
			gologger.Info().Msg("Url pending queue is empty, break")
			break
		}
		req, ok := c.pendingQueue.Pop()
		if !ok {
			continue
		}
		wg.Add()
		atomic.AddInt32(&running, 1)

		go func() {
			defer wg.Done()
			defer atomic.AddInt32(&running, -1)

			resp, err := c.navigateRequest(browserInstance, req)
			if err != nil {
				gologger.Error().Msg(err.Error())
				return
			}
			if resp == nil {
				return
			}
			parser.ParseResponse(*resp, callback)
		}()
	}

	wg.Wait()

	if err := c.outputWriter.Close(); err != nil {
		gologger.Error().Msgf("Save file error %s", err.Error())
	}

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
	page, err := browser.Page(proto.TargetCreateTarget{URL: strings.Join([]string{req.Url}, "/")})
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	defer page.Close()

	//https://github.com/go-rod/rod/issues/230
	err = page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent:      UserAgent,
		AcceptLanguage: Language,
	})
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	page = page.Timeout(time.Duration(c.option.Timeout) * time.Second)

	lastTimestamp := time.Now().Unix()

	requestMap := sync.Map{}

	go page.EachEvent(func(e *proto.NetworkLoadingFinished) {
		defer func() {
			lastTimestamp = time.Now().Unix()
		}()

		data, ok := requestMap.Load(e.RequestID)
		if !ok {
			gologger.Debug().Msg("RequestID not exist, skip")
			return
		}
		event := data.(*types.EventListen)
		request := event.Request
		response := event.Response

		gologger.Debug().Msgf("Start process RequestID %s", e.RequestID)

		if request == nil {
			request = &proto.NetworkRequest{
				Method: http.MethodGet,
				URL:    response.URL,
			}
		}
		_url := request.URL

		urlParsed, _ := url.Parse(_url)
		if urlParsed.Host != c.rootHost {
			gologger.Debug().Msgf("out of site url %s, skip", _url)
			return // 外部站点
		}

		go func() {
			defer func() {
				lastTimestamp = time.Now().Unix()
			}()
			m := proto.NetworkGetResponseBody{RequestID: e.RequestID}
			r, err := m.Call(page)
			if err != nil {
				gologger.Error().Msgf("GetResponseBody error: %s", err.Error())
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
				ResponseContentType: contentType,
				Depth:               req.Depth + 1,
			}

			var expandResult = c.expandClient.Run(_url, contentType)
			if len(expandResult) == 0 {
				c.saveResponse(respResult)
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
		}()

	}, func(e *proto.NetworkRequestWillBeSent) {
		requestMap.Store(e.RequestID, &types.EventListen{
			RequestId: e.RequestID,
			Request:   e.Request,
		})
	}, func(e *proto.NetworkRequestWillBeSentExtraInfo) {
		//https://github.com/go-rod/rod/issues/351 目前没需求
		requestMap.Store(e.RequestID, &types.EventListen{
			RequestId: e.RequestID,
		})
		gologger.Debug().Msgf("Add new request_id %s", e.RequestID)
	}, func(e *proto.NetworkResponseReceived) {
		gologger.Debug().Msgf("Get response for request_id %s", e.RequestID)
		data, ok := requestMap.Load(e.RequestID)
		if !ok {
			requestMap.Store(e.RequestID, &types.EventListen{
				RequestId: e.RequestID,
				Response:  e.Response,
			})
		} else {
			event := data.(*types.EventListen)
			event.Response = e.Response
		}
	}, func(e *proto.PageJavascriptDialogOpening) {
		if e.Type == proto.PageDialogTypeAlert {
			d := proto.PageHandleJavaScriptDialog{
				Accept:     true,
				PromptText: "",
			}
			err := d.Call(page)
			if err != nil {
				gologger.Error().Msgf("Closed PageDialogTypeAlert error: %s", err.Error())
			}
		}

	})()

	err = rod.Try(func() {
		page.MustWaitNavigation()
		page.MustWaitLoad()
		page.MustWaitIdle()
		page.MustReload() //reload page
	})
	if err != nil {
		return nil, errors.Wrap(err, "wait load error")
	}

	c.waitLoaded(lastTimestamp, 5)

	html, err := page.HTML()
	if err != nil {
		return nil, errors.Wrap(err, "could not get html")
	}

	if utils.GetUrlPath(req.Url) == utils.GetUrlPath(c.option.Url) {
		_ = rod.Try(func() {
			page.MustScreenshotFullPage(filepath.Join(c.targetDir, "screenshot", urlutil.GetUrlHost(req.Url)+".png"))
		})
	}

	c.processLoginForm(page)

	//一些请求是 解析 js 加载的，这里设置长一点
	c.waitLoaded(lastTimestamp, 30)

	return &types.Response{
		Body: html,
	}, nil
}

func (c *Crawler) waitLoaded(timestamp int64, interval int64) {
	gologger.Info().Msgf("Start wait loading... %d s", interval)
	for {
		if time.Now().Unix()-timestamp > interval {
			break
		}
		time.Sleep(time.Millisecond * 300)
	}
}

func (c *Crawler) processLoginForm(page *rod.Page) {
	forms, err := page.ElementsX("//form")
	if err != nil {
		return
	}

	gologger.Info().Msgf("find form size %d", len(forms))
	for _, formElement := range forms {
		inputs, err := formElement.ElementsX("//input")
		if err != nil {
			continue
		}
		if len(inputs) == 0 {
			continue
		}

		gologger.Info().Msgf("Input element size %d", len(inputs))
		for _, inputElement := range inputs {
			if !c.attributeMock.IsEnable(inputElement) {
				gologger.Info().Msgf("<%s> is dis enable, skip", inputElement.String())
				continue
			}
			v := c.attributeMock.MockValue(inputElement)
			if v == "" {
				if c.attributeMock.IsLoginBtn(inputElement) {
					break
				}
				gologger.Info().Msgf("MockValue is Empty, skip, %s", inputElement.String())
				continue
			}
			gologger.Info().Msgf("Start input <%s> for %s", v, inputElement.String())
			err := inputElement.Input(v)
			if err != nil {
				gologger.Error().Msgf("Type Input value error %s", err.Error())
			}
			time.Sleep(1)
		}

		for _, loginXpath := range c.attributeMock.LoginXpaths {
			loginBtn, err := formElement.ElementX(loginXpath)
			if err != nil {
				continue
			}
			gologger.Info().Msgf("Find login btn %s", loginBtn)
			err = loginBtn.Click(proto.InputMouseButtonLeft, 1)
			if err != nil {
				gologger.Error().Msgf("Btn click error %s", err.Error())
			}
			break
		}
	}
}

func (c *Crawler) saveResponse(result *types.ResponseResult) {
	c.crawledUrl.Add(result.Url) //backup
	err := c.outputWriter.Write(*result)
	if err != nil {
		gologger.Error().Msgf("Write response error %s", err.Error())
	}

	c.saveFile(utils.GetUrlPath(result.Url), result)

	//如果是入口页面，保存一份 index.html
	if utils.IsSameURL(result.Url, c.option.Url) && result.ResponseContentType == "text/html" {
		parsed, _ := url.Parse(result.Url)
		result.Url = fmt.Sprintf(`%s://%s`, parsed.Scheme, parsed.Host)
		err = c.outputWriter.Write(*result)
		if err != nil {
			gologger.Error().Msgf("Write response error %s", err.Error())
		}
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
	gologger.Debug().Msgf("Save file %s", urlPath)
	var data interface{}
	data = resp.Body

	urlPath = strings.TrimRight(urlPath, "/")

	var paths []string
	if urlPath == "" {
		paths = []string{c.targetDir, "index.html"}
	} else {
		paths = []string{c.targetDir, urlPath}
		//https://github.com/yangyang5214/clone-alive/issues/15
		parts := strings.Split(urlPath, "/")
		lastPath := parts[len(parts)-1]
		if !strings.Contains(lastPath, ".") {
			fileNameSuffix := types.ConvertFileName(resp.ResponseContentType)
			paths = append(paths, "index."+fileNameSuffix)
		}
	}

	if resp.ResponseContentType == types.TextHtml {
		//replace original url
		for _, item := range c.domains {
			data = strings.Replace(resp.Body, item, "", -1)
		}
		data = strings.Replace(resp.Body, c.rootHost, "", -1)
		data = strings.Replace(resp.Body, strings.Split(c.rootHost, ":")[0], "", -1)

		//replace input
		data = strings.Replace(resp.Body, magic.DefaultEmail, "", -1)
		data = strings.Replace(resp.Body, magic.DefaultUser, "", -1)
		data = strings.Replace(resp.Body, magic.DefaultText, "", -1)
	} else if strings.HasPrefix(resp.ResponseContentType, "image/") {
		data = base64.NewDecoder(base64.StdEncoding, strings.NewReader(data.(string)))
	}

	p := filepath.Join(paths...)
	if fileutil.FileExists(p) {
		gologger.Info().Msgf("File %s is exists, skip", p)
		return
	}
	err := rod_util.OutputFile(p, data)
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
