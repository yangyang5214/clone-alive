package types

import (
	"github.com/yangyang5214/clone-alive/pkg/utils"
	"strings"
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
	fileType := utils.GetSplitLast(lastPath, ".")
	_, ok := FileType[strings.ToLower(fileType)]
	if ok {
		return true
	}
	return false
}
