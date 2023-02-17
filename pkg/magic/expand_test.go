package magic

import (
	"testing"
)

func TestGenExpand(t *testing.T) {
	expand := NewExpand()
	r := expand.Run("https://120.27.184.164/?module=captcha&0.09322127984833917", "image/png")
	t.Log(len(r))
}
