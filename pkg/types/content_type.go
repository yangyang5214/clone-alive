package types

import "strings"

type ContentType = string

const (
	TextHtml                  ContentType = "text/html"
	ImagePng                  ContentType = "image/png"
	ImageJpeg                 ContentType = "image/jpeg"
	ApplicationJson           ContentType = "application/json"
	TextJs                    ContentType = "text/javascript"
	TextCss                   ContentType = "text/css"
	ApplicationJsonUtf8       ContentType = "application/json;charset=utf-8"
	ApplicationFormUrlencoded ContentType = "application/x-www-form-urlencoded"
)

func ConvertContentType(contentType ContentType) string {
	switch contentType {
	case "", "<nil>":
		return ""
	case ApplicationFormUrlencoded:
		return ""
	case ApplicationJsonUtf8:
		return ApplicationJson
	case TextJs:
		return "application/javascript"
	default:
		return contentType
	}
}

func ConvertFileName(contentType ContentType) string {
	splits := strings.Split(contentType, "/")
	if len(splits) == 2 {
		return splits[1]
	}
	return ""
}
