package magic

import (
	"strings"

	"github.com/go-rod/rod"
	"github.com/projectdiscovery/gologger"
)

type AttributeType = string

const (
	DefaultEmail = "beer@beer.com"
	DefaultUser  = "clone-alive"
	DefaultText  = "Clone-Alive_magic123~"
)

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
	Logins      []string
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

func (a *Attribute) IsEnable(el *rod.Element) bool {
	_, err := el.ElementByJS(rod.Eval(`() => !this.disabled`))
	if err != nil {
		return true
	}
	return false
}

func (a *Attribute) IsLoginBtn(element *rod.Element) bool {
	for _, item := range a.Logins {
		v := a.MustAttribute(element, item)
		if v == "" {
			continue
		}
		v = strings.Replace(v, " ", "", -1)
		gologger.Info().Msgf("Attribute <%s> is %s", item, v)
		if v == "登录" {
			return true
		}
		if strings.Contains(v, "loginbtn") {
			return true
		}
		if strings.Contains(v, "form.submit()") {
			return true
		}
		if v == "button" {
			return true
		}
	}
	return false
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
			return DefaultEmail
		case user, login:
			return DefaultUser
		case password, text, queryString:
			return DefaultText
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
		Logins: []string{"value", "id", "class", "onclick", "type"},
		LoginXpaths: []string{
			"//button//*[contains(text(),'Sign in')]", //http://47.93.32.144:3000/auth/login
			"//*[@id='btnLogin']",                     //http://58.56.78.6:81/pages/login.jsp
			"//*[@type='button']",                     //https://217.181.140.91:4443/
			"//*[@type='submit']",                     //http://106.52.194.58:8090/login.action
			"//*[@id='loginBtn']",
			"//*[@id='btn_log']",
			"//*[@class='loginbtn']",
			"//*[contains(@class,'loginbtn')]",
			"//*[contains(@value,'登 录')]",
			"//*[@id=\"signIn\"]", //https://202.3.166.101/SAAS/auth/login
			"//*[@value=\"Sign In\"]",
		},
	}
}
