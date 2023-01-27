package parser

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/yangyang5214/clone-alive/pkg/types"
	"strings"
	"testing"
)

func buildResponse() types.Response {
	resp := types.Response{
		Body: "<a class=\"post-card-content-link\" href=\"/github-xiang-mu-zhi-chi-go-install/\">\n            <header class=\"post-card-header\">\n                <div class=\"post-card-tags\">\n                        <span class=\"post-card-primary-tag\">Golang</span>\n                </div>\n                <h2 class=\"post-card-title\">\n                    github 项目支持 go install\n                </h2>\n            </header>\n                <div class=\"post-card-excerpt\">自己写了个项目，支持别人直接 go install xxx\n\n\n\n实际就是 go mod name 使用 GitHub 前缀\n\n\n * 初始化\n\n\n# demo\ngo mod init example.com/greetings\n\n# 自己例子\ngo mod init github.com/yangyang5214/clone-alive\n\n\n\n\n * 已存在\n\n\n直接找到 go.mod 文件，全局 rename module name\n\ngoland: shift + f6\n\n\n\n例子：https://github.com/yangyang5214/clone-alive/</div>\n        </a>",
	}
	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(resp.Body))
	resp.Reader = doc
	return resp
}

func TestBodyATagParser(t *testing.T) {
	resp := buildResponse()
	result := bodyATagParser(resp)
	t.Log(result)
}
