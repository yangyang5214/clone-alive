package magic

import (
	"github.com/yangyang5214/clone-alive/pkg/utils"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
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

func Hit(urlPath string, contentType string) bool {
	if !strings.HasPrefix(contentType, "image") {
		return false
	}
	if strings.Contains(urlPath, "verifycode") {
		return true
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
		httpClient: &http.Client{},
		retryCount: RetryCount,
	}
}
