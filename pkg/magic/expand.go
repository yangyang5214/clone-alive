package magic

import (
	"github.com/projectdiscovery/gologger"
	"github.com/yangyang5214/clone-alive/pkg/types"
	"github.com/yangyang5214/clone-alive/pkg/utils"
	"github.com/yangyang5214/gou/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

const RetryCount = 30

type ExpandVerifyCode struct {
	httpClient *httputil.HttpClient
	retryCount int
}

type VerifyCodeResults struct {
	UrlStr string
	Body   string
}

var partUrlPath = []string{
	"verifycode", //http://58.56.78.6:81/pages/login.jsp
	"getCode",
	"servlets/vms",   //http://58.250.50.115:5050/
	"login/code",     //http://10.0.81.29:8001/
	"module=captcha", //https://120.27.184.164/
	"createcode",     //https://222.187.115.230:10443/
}

func Hit(urlPath string) bool {
	for _, item := range partUrlPath {
		if strings.Contains(urlPath, item) {
			gologger.Info().Msgf("Url <%s> Hit <%s>", urlPath, item)
			return true
		}
	}
	return false
}

func RebuildUrl(urlpath string, index int, contentType string) string {
	if contentType == "" {
		contentType = types.ImagePng
	}
	return path.Join(utils.GetRealUrl(urlpath), strconv.Itoa(index)+"."+utils.GetSplitLast(contentType, "/"))
}

func (e *ExpandVerifyCode) Run(urlStr string, contentType string) []*VerifyCodeResults {
	urlParsed, _ := url.Parse(urlStr)
	if !Hit(urlParsed.String()) {
		return nil
	}

	var result []*VerifyCodeResults

	for i := 0; i < e.retryCount; i++ {
		resp, err := e.httpClient.Get(urlStr)
		if err != nil {
			break
		}
		respBody, err := e.httpClient.ReadBody(resp)
		if err != nil {
			break
		}
		if respBody == nil {
			gologger.Error().Msgf("Fetch http req failed: %s", urlStr)
			continue
		}
		result = append(result, &VerifyCodeResults{
			UrlStr: RebuildUrl(urlParsed.Path, i, contentType),
			Body:   string(respBody),
		})
	}
	return result
}

func NewExpand() *ExpandVerifyCode {
	return &ExpandVerifyCode{
		httpClient: httputil.NewClient(httputil.DefaultOptions),
		retryCount: RetryCount,
	}
}
