package utils

import (
	"fmt"
	"github.com/projectdiscovery/gologger"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func IsURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

func GetUrlPath(u string) string {
	urlParsed, _ := url.Parse(u)
	p := urlParsed.Path
	if p == "" {
		p = "/"
	}
	return p
}

func GetUrlHost(u string) string {
	urlParsed, _ := url.Parse(u)
	return urlParsed.Host
}

func GetDomain(u string) string {
	urlParsed, _ := url.Parse(u)
	return fmt.Sprintf("%s://%s", urlParsed.Scheme, urlParsed.Host)
}

func IsSameURL(u1 string, u2 string) bool {
	return GetUrlPath(u1) == GetUrlPath(u2)
}

func DoHttpReq(req *http.Request, httpClient *http.Client) []byte {
	resp, err := httpClient.Do(req)
	if err != nil {
		gologger.Error().Msgf("DoHttpReq error %s", err.Error())
		return nil
	}
	if resp.StatusCode != 200 {
		gologger.Error().Msgf("DoHttpReq error StatusCode %d", resp.StatusCode)
		return nil
	}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	return bytes
}

func GetSplitLast(str string, seq string) string {
	r := strings.Split(str, seq)
	length := len(r)
	if length == 0 {
		return str
	}
	return r[length-1]
}
