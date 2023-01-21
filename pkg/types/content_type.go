package types

type ContentType = string

const (
	TextHtml        ContentType = "text/html"
	ImagePng        ContentType = "image/png"
	ImageJpeg       ContentType = "image/jpeg"
	ApplicationJson ContentType = "application/json"
	TextJs          ContentType = "text/javascript"
	TextCss         ContentType = "text/css"
)

func ConvertContentType(contentType ContentType) string {
	switch contentType {
	case TextJs:
		return "application/javascript"
	default:
		return contentType
	}
}
