package utils

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestUrlParse(t *testing.T) {
	r := GetUrlPath("http://58.221.18.142:444/current_config/preLanguage?16766026183290026_dc=1676602618329")
	assert.Equal(t, r, "/current_config/preLanguage")

	assert.Equal(t, GetUrlPath("https://www.baidu.com"), "/")
	assert.Equal(t, GetUrlPath("https://www.baidu.com/"), "/")
	assert.Equal(t, GetUrlPath("http://10.0.81.29:3001/"), "/")
}

func TestSplit(t *testing.T) {
	t.Log(strings.Split("127.0.0.1:9090", ":")[0])
}

func TestGetRealUrl(t *testing.T) {
	t.Log(GetRealUrl("https://183.134.103.232/resources/jquery/jquery.min.js%3Fjs_ver=2023-02-16"))
}

func TestIsSameURL(t *testing.T) {
	r := IsSameURL("https://120.27.184.164/?module=captcha&0.09005321683750123", "https://120.27.184.164/?module=captcha")
	t.Log(r)
}
