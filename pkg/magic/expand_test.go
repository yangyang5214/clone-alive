package magic

import (
	"testing"

	"github.com/kr/pretty"
)

func TestGenExpand(t *testing.T) {
	expand := NewExpand(3)

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
