package magic

import (
	"crypto/tls"
	"github.com/projectdiscovery/gologger"
	"github.com/yangyang5214/clone-alive/pkg/utils"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

const RetryCount = 30

type ExpandVerifyCode struct {
	httpClient *http.Client
	retryCount int
}

type VerifyCodeResults struct {
	UrlStr string
	Body   string
}

var partUrlPath = []string{
	"verifycode", //http://58.56.78.6:81/pages/login.jsp
	"getCode",
	"servlets/vms", //http://58.250.50.115:5050/
	"login/code",   //http://10.0.81.29:8001/
}

func Hit(urlPath string, contentType string) bool {
	for _, item := range partUrlPath {
		if strings.Contains(urlPath, item) {
			return true
		}
	}
	return false
}

func RebuildUrl(urlpath string, index int, contentType string) string {
	return path.Join(urlpath, strconv.Itoa(index)+"."+strings.Split(contentType, "/")[1])
}

func (e *ExpandVerifyCode) Run(urlStr string, contentType string) []*VerifyCodeResults {
	urlParsed, _ := url.Parse(urlStr)
	if !Hit(urlParsed.Path, contentType) {
		return nil
	}
	req := &http.Request{
		Method: "GET",
		URL:    urlParsed,
	}

	var result []*VerifyCodeResults

	for i := 0; i < e.retryCount; i++ {
		respBody := utils.DoHttpReq(req, e.httpClient)
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
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &ExpandVerifyCode{
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   60 * time.Second},
		retryCount: RetryCount,
	}
}
