package utils

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func IsURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

func GetUrlPath(u string) string {
	urlParsed, _ := url.Parse(u)
	return urlParsed.Path
}

func IsSameURL(u1 string, u2 string) bool {
	return GetUrlPath(u1) == GetUrlPath(u2)
}

func DoHttpReq(req *http.Request, httpClient *http.Client) []byte {
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil
	}
	if resp.StatusCode != 200 {
		return nil
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	return bytes
}
