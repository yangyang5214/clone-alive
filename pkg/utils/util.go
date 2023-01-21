package utils

import (
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
