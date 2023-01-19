package utils

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestUrlParse(t *testing.T) {
	p, _ := url.Parse("https://www.baidu.com")
	assert.Equal(t, p.Path, "")

	p1, _ := url.Parse("https://www.baidu.com/")
	assert.Equal(t, p1.Path, "/")
}
