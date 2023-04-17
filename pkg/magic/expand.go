package magic

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/projectdiscovery/gologger"
	"github.com/yangyang5214/clone-alive/pkg/types"
	"github.com/yangyang5214/clone-alive/pkg/utils"
	fileutil "github.com/yangyang5214/gou/file"
	"github.com/yangyang5214/gou/set"
)

const (
	RetryCount = 30
)

type ExpandVerifyCode struct {
	httpClient   *http.Client
	retryCount   int
	partUrlPaths *set.Set[string]
}

func NewExpand(retry int, verifyCodePath string) *ExpandVerifyCode {
	return &ExpandVerifyCode{
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: 10 * time.Second,
		},
		retryCount:   retry,
		partUrlPaths: fileutil.FileReadLinesSet(verifyCodePath),
	}
}

type VerifyCodeResults struct {
	UrlStr string
	Body   string
}

func Hit(urlPath string, partUrlPaths *set.Set[string]) bool {
	if partUrlPaths.Contains(urlPath) {
		gologger.Info().Msgf("Url <%s> Hit <%s>", urlPath, urlPath)
		return true
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

	if !Hit(urlParsed.String(), e.partUrlPaths) {
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
