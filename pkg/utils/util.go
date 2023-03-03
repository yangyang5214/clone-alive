package utils

import (
	"fmt"
	"net/url"
	"strings"
)

func IsURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

func GetUrlPath(u string) string {
	urlParsed, _ := url.Parse(GetRealUrl(u))
	p := urlParsed.Path
	if p == "" {
		p = "/"
	}
	return p
}

func GetDomain(u string) string {
	urlParsed, _ := url.Parse(u)
	return fmt.Sprintf("%s://%s", urlParsed.Scheme, urlParsed.Host)
}

func GetDomains(u string) []string {
	urlParsed, _ := url.Parse(u)
	return []string{
		fmt.Sprintf("%s://%s", "http", urlParsed.Host+":80"),
		fmt.Sprintf("%s://%s", "https", urlParsed.Host+":443"),
		fmt.Sprintf("%s://%s", "https", urlParsed.Host),
		fmt.Sprintf("%s://%s", "https", GetSplitLast(urlParsed.Host, ":")),
		fmt.Sprintf("%s://%s", "http", urlParsed.Host),
		fmt.Sprintf("%s://%s", "http", GetSplitLast(urlParsed.Host, ":")),
	}
}

// GetRealUrl is remove query
func GetRealUrl(u string) string {
	r := strings.Split(u, "?")
	if len(r) == 1 {
		r = strings.Split(u, "%3F")
	}
	return r[0]
}

// IsSameURL todo with param
func IsSameURL(u1 string, u2 string) bool {
	if u1 == u2 {
		return true
	}
	return GetUrlPath(u1) == GetUrlPath(u2)
}

func GetSplitLast(str string, seq string) string {
	r := strings.Split(str, seq)
	length := len(r)
	if length == 0 {
		return str
	}
	return r[length-1]
}
