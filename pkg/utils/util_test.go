package utils

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"strings"
	"testing"
)

func TestUrlParse(t *testing.T) {
	p, _ := url.Parse("https://www.baidu.com")
	assert.Equal(t, p.Path, "")

	p1, _ := url.Parse("https://www.baidu.com/")
	assert.Equal(t, p1.Path, "/")
	t.Log(p1.Host)

	p3, _ := url.Parse("http://10.0.81.29:3001/")
	t.Log(p3.Host)
	t.Log(p3.Hostname())
}

func TestSplit(t *testing.T) {
	t.Log(strings.Split("127.0.0.1:9090", ":")[0])
}
