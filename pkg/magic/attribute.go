package magic

import (
	"github.com/go-rod/rod"
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

type Attribute struct {
	Inputs      []string
	LoginXpaths []string
}

func (a *Attribute) MustAttribute(el *rod.Element, name string) string {
	attr, err := el.Attribute(name)
	if err != nil {
		return ""
	}
	if attr == nil {
		return ""
	}
	return *attr
}

func (a *Attribute) MockValue(element *rod.Element) string {

	for _, item := range a.Inputs {
		attribute := a.MustAttribute(element, item)
		if attribute == "" {
			continue
		}
		typeStr := strings.ToLower(attribute)

		//id="userName"
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
			gologger.Warning().Msgf("Find New Attribute type %s", typeStr)
		}
	}
	return ""
}

func NewAttribute() *Attribute {
	return &Attribute{
		//if type == 'hidden', skip first
		Inputs: []string{"type", "id", "name", "placeholder", "ng-model"},
		LoginXpaths: []string{
			"//button//*[contains(text(),'Sign in')]", //http://47.93.32.144:3000/auth/login
			"//*[@id='btnLogin']",                     //http://58.56.78.6:81/pages/login.jsp
			"//*[@type='button']",                     //https://217.181.140.91:4443/
			"//*[@type='submit']",                     //http://106.52.194.58:8090/login.action
			"//*[@id='loginBtn']",
		},
	}
}
