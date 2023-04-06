package types

import (
	"strings"

	"github.com/yangyang5214/clone-alive/pkg/utils"
)

var FileType = map[string]string{
	"png":  "",
	"jpeg": "",
	"css":  "",
	"js":   "",
}

func IsStaticFile(url string) bool {
	url = utils.GetRealUrl(url)
	lastPath := utils.GetSplitLast(url, "/")

	// SAAS/jersey/manager/api/images/5101/
	if !strings.Contains(lastPath, ".") {
		return false
	}
	fileType := utils.GetSplitLast(lastPath, ".")
	_, ok := FileType[strings.ToLower(fileType)]
	if ok {
		return true
	}
	return false
}
