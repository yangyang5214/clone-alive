package magic

import (
	"crypto/tls"
	"github.com/projectdiscovery/gologger"
	"github.com/yangyang5214/clone-alive/pkg/types"
	"github.com/yangyang5214/clone-alive/pkg/utils"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	RetryCount = 30
)

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

func (e *ExpandVerifyCode) Run(urlStr string, contentType string) (result []*VerifyCodeResults) {
	//skip error url
	urlParsed, err := url.Parse(urlStr)
	if err != nil {
		return []*VerifyCodeResults{}
	}

	if !Hit(urlParsed.String()) {
		gologger.Debug().Msgf("Url <%s> not match rules, skip", urlStr)
		return []*VerifyCodeResults{}
	}

	for i := 0; i < e.retryCount; i++ {
		result = append(result, e.httpGet(urlParsed, i, contentType))
	}
	return result
}

func (e *ExpandVerifyCode) httpGet(u *url.URL, index int, contentType string) *VerifyCodeResults {
	resp, err := e.httpClient.Get(u.String())
	defer resp.Body.Close()
	if err != nil {
		gologger.Error().Msgf("Http error %s", err.Error())
		return nil
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		gologger.Error().Msgf("Http body read error %s", err.Error())
		return nil
	}
	return &VerifyCodeResults{
		UrlStr: RebuildUrl(u.Path, index, contentType),
		Body:   string(respBody),
	}
}

func NewExpand(retry int) *ExpandVerifyCode {
	return &ExpandVerifyCode{
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: 10 * time.Second,
		},
		retryCount: retry,
	}
}
