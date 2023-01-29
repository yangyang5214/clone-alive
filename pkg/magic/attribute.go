package magic

import "github.com/projectdiscovery/gologger"

type AttributeType = string

const (
	email    AttributeType = "email"
	password AttributeType = "password"
	text     AttributeType = "text"
	hidden   AttributeType = "hidden"
	INPUT    AttributeType = "input"
	user     AttributeType = "user"
)

var LoginXpaths = []string{
	"//button//*[contains(text(),'Sign in')]", //http://47.93.32.144:3000/auth/login
	"//*[contains(text(),'登录')]",              //http://58.56.78.6:81/pages/login.jsp
	"//*[@type='button']",                     //https://217.181.140.91:4443/
	"//*[@id='loginBtn']",                     //https://121.22.125.130:4433/toLogin?forceLogout=1
}

type Attribute struct {
}

func (a *Attribute) MockValue(typeStr string) string {
	switch typeStr {
	case user:
		return "clone-alive"
	case hidden:
		return ""
	case email:
		return "beer@beer.com"
	case password, text:
		return "Clone-Alive_magic123~"
	default:
		gologger.Error().Msgf("Find New Attribute type %s", typeStr)
	}
	return "Clone-Alive_magic123~"
}

func NewAttribute() *Attribute {
	return &Attribute{}
}
