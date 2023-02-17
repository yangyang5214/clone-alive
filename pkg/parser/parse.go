package parser

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/projectdiscovery/gologger"
	"github.com/yangyang5214/clone-alive/pkg/types"
	"strings"
)

type ResponseParserFunc func(resp types.Response) []string

var parsers = []ResponseParserFunc{
	bodyATagParser,
	bodyScriptSrcTagParser,
	bodyLinkHrefTagParser,
	bodyImgSrcTagParser,
	//bodyLinkHrefTagParser,
	//bodyEmbedTagParser,
	//bodyFrameTagParser,
	//bodyIframeTagParser,
	//bodyInputSrcTagParser,
	//bodyIsindexActionTagParser,
	//bodyScriptSrcTagParser,
	//bodyBackgroundTagParser,
	//bodyImgTagParser,
}

// ParseResponse  from https://github.com/projectdiscovery/katana/blob/main/pkg/engine/parser/parser.go
func ParseResponse(resp types.Response, callback func(req types.Request)) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.Body))
	if err != nil {
		gologger.Error().Msgf("string to doc error %s", err.Error())
		return
	}
	resp.Reader = doc

	for _, parseFun := range parsers {
		for _, urlStr := range parseFun(resp) {
			callback(types.Request{
				Url:   urlStr,
				Depth: resp.Depth,
			})
		}
	}

}

// bodyATagParser parses A tag from response
func bodyATagParser(resp types.Response) (urls []string) {
	resp.Reader.Find("a").Each(func(i int, item *goquery.Selection) {
		href, ok := item.Attr("href")
		if ok && href != "" {
			urls = append(urls, href)
		}
		ping, ok := item.Attr("ping")
		if ok && ping != "" {
			urls = append(urls, ping)
		}
	})
	return urls
}

// bodyScriptSrcTagParser parses script src tag from response
func bodyScriptSrcTagParser(resp types.Response) (urls []string) {
	resp.Reader.Find("script[src]").Each(func(i int, item *goquery.Selection) {
		src, ok := item.Attr("src")
		if ok && src != "" {
			urls = append(urls, src)
		}
	})
	return urls
}

// bodyImgSrcTagParser… parses script src tag from response
func bodyImgSrcTagParser(resp types.Response) (urls []string) {
	resp.Reader.Find("img[src]").Each(func(i int, item *goquery.Selection) {
		src, ok := item.Attr("src")
		if ok && src != "" {
			urls = append(urls, src)
		}
	})
	return urls
}

// bodyScriptSrcTagParser parses script src tag from response
func bodyLinkHrefTagParser(resp types.Response) (urls []string) {
	resp.Reader.Find("link[href]").Each(func(i int, item *goquery.Selection) {
		src, ok := item.Attr("href")
		if ok && src != "" {
			urls = append(urls, src)
		}
	})
	return urls
}

//// bodyEmbedTagParser parses Embed tag from response
//func bodyEmbedTagParser(resp types.Response, callback func(types.Request)) {
//	resp.Reader.Find("embed[src]").Each(func(i int, item *goquery.Selection) {
//		src, ok := item.Attr("src")
//		if ok && src != "" {
//			callback(types.Request{
//				Url:   src,
//				Depth: resp.Depth,
//			})
//		}
//	})
//}
//
//// bodyFrameTagParser parses frame tag from response
//func bodyFrameTagParser(resp types.Response, callback func(types.Request)) {
//	resp.Reader.Find("frame[src]").Each(func(i int, item *goquery.Selection) {
//		src, ok := item.Attr("src")
//		if ok && src != "" {
//			callback(types.Request{
//				Url:   src,
//				Depth: resp.Depth,
//			})
//		}
//	})
//}
//
//// bodyIframeTagParser parses iframe tag from response
//func bodyIframeTagParser(resp types.Response, callback func(types.Request)) {
//	resp.Reader.Find("iframe").Each(func(i int, item *goquery.Selection) {
//		src, ok := item.Attr("src")
//		if ok && src != "" {
//			callback(types.Request{
//				Url:   src,
//				Depth: resp.Depth,
//			})
//		}
//		srcDoc, ok := item.Attr("srcdoc")
//		if ok && srcDoc != "" {
//			endpoints := utils.ExtractRelativeEndpoints(srcDoc)
//			for _, endpoint := range endpoints {
//				callback(types.Request{
//					Url:   endpoint,
//					Depth: resp.Depth,
//				})
//			}
//		}
//	})
//}
//
//// bodyInputSrcTagParser parses input image src tag from response
//func bodyInputSrcTagParser(resp types.Response, callback func(types.Request)) {
//	resp.Reader.Find("input[type='image' i]").Each(func(i int, item *goquery.Selection) {
//		src, ok := item.Attr("src")
//		if ok && src != "" {
//			callback(types.Request{
//				Url:   src,
//				Depth: resp.Depth,
//			})
//		}
//	})
//}
//
//// bodyIsindexActionTagParser parses isindex action tag from response
//func bodyIsindexActionTagParser(resp types.Response, callback func(types.Request)) {
//	resp.Reader.Find("isindex[action]").Each(func(i int, item *goquery.Selection) {
//		src, ok := item.Attr("action")
//		if ok && src != "" {
//			callback(types.Request{
//				Url:   src,
//				Depth: resp.Depth,
//			})
//		}
//	})
//}
//

//// bodyBackgroundTagParser parses body background tag from response
//func bodyBackgroundTagParser(resp types.Response, callback func(types.Request)) {
//	resp.Reader.Find("body[background]").Each(func(i int, item *goquery.Selection) {
//		src, ok := item.Attr("background")
//		if ok && src != "" {
//			callback(types.Request{
//				Url:   src,
//				Depth: resp.Depth,
//			})
//		}
//	})
//}
//
//func bodyImgTagParser(resp types.Response, callback func(types.Request)) {
//	resp.Reader.Find("img").Each(func(i int, item *goquery.Selection) {
//		srcDynsrc, ok := item.Attr("dynsrc")
//		if ok && srcDynsrc != "" {
//			callback(types.Request{
//				Url:   srcDynsrc,
//				Depth: resp.Depth,
//			})
//		}
//		srcLongdesc, ok := item.Attr("longdesc")
//		if ok && srcLongdesc != "" {
//			callback(types.Request{
//				Url:   srcLongdesc,
//				Depth: resp.Depth,
//			})
//		}
//		srcLowsrc, ok := item.Attr("lowsrc")
//		if ok && srcLowsrc != "" {
//			callback(types.Request{
//				Url:   srcLowsrc,
//				Depth: resp.Depth,
//			})
//		}
//		src, ok := item.Attr("src")
//		if ok && src != "" && src != "#" {
//			if strings.HasPrefix(src, "data:") {
//				// TODO: Add data:uri/data:image parsing
//				return
//			}
//			callback(types.Request{
//				Depth: resp.Depth,
//
//				Url: src,
//			})
//		}
//		srcSet, ok := item.Attr("srcset")
//		if ok && srcSet != "" {
//			for _, value := range utils.ParseSRCSetTag(srcSet) {
//				callback(types.Request{
//					Url:   value,
//					Depth: resp.Depth,
//				})
//			}
//		}
//	})
//}
