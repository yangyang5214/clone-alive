package magic

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kr/pretty"
)

var expand *ExpandVerifyCode

func init() {
	expand = NewExpand(3, "./../../config/verify_code")
}

func TestGenExpand(t *testing.T) {

	t.Run("case1", func(t *testing.T) {
		r := expand.Run("https://120.27.184.164/?module=captcha&0.09322127984833917", "image/png")
		t.Log(len(r))
	})

	t.Run("case2", func(t *testing.T) {
		u := "http://58.56.78.6:81/verifycode.do?width=70&height=20&codecount=4&codestyle=digit&timestamp=1679123449408"
		r := expand.Run(u, "image/png")
		t.Log(len(r))
		pretty.Log(r)
	})

}

func TestHit(t *testing.T) {
	assert.Equal(t, Hit("http://127.0.0.1:8080/verifycode.do?width=70&height=20&codecount=4&codestyle=digit&timestamp=1679123449408", expand.partUrlPaths), true)
}
