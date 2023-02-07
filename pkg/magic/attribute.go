package magic

import (
	"github.com/projectdiscovery/gologger"
	"strings"
)

type AttributeType = string

const (
	email       AttributeType = "email"
	password    AttributeType = "password"
	text        AttributeType = "text"
	hidden      AttributeType = "hidden"
	login       AttributeType = "login"
	input       AttributeType = "input"
	user        AttributeType = "user"
	checkbox    AttributeType = "checkbox"
	submit      AttributeType = "submit"
	queryString AttributeType = "queryString"
)

var LoginXpaths = []string{
	"//button//*[contains(text(),'Sign in')]", //http://47.93.32.144:3000/auth/login
	"//*[contains(text(),'登录')]",              //http://58.56.78.6:81/pages/login.jsp
	"//*[@type='button']",                     //https://217.181.140.91:4443/
	"//*[@type='submit']",                     //http://106.52.194.58:8090/login.action
	"//*[@id='loginBtn']",                     //https://121.22.125.130:4433/toLogin?forceLogout=1
}

var ignoreMap = map[string]bool{
	input: true,
	text:  true,
}

type Attribute struct {
}

func (a *Attribute) IsIgnore(typeStr string) bool {
	_, ok := ignoreMap[typeStr]
	return ok
}

func (a *Attribute) MockValue(typeStr string) string {
	if strings.Contains(typeStr, "user") {
		typeStr = user
	}
	switch typeStr {
	case checkbox, submit, hidden:
		return ""
	case email:
		return "beer@beer.com"
	case user, login:
		return "clone-alive"
	case password, text, queryString:
		return "Clone-Alive_magic123~"
	default:
		gologger.Info().Msgf("Find New Attribute type %s", typeStr)
	}
	return "Clone-Alive_magic123~"
}

func NewAttribute() *Attribute {
	return &Attribute{}
}
