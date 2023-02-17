package alive

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/gologger"
	"github.com/yangyang5214/clone-alive/pkg/magic"
	"github.com/yangyang5214/clone-alive/pkg/types"
	"github.com/yangyang5214/clone-alive/pkg/utils"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

type Alive struct {
	option   types.AliveOption
	routeMap sync.Map
}

type RouteResp struct {
	Path string
	Resp *types.ResponseResult
}

func New(option types.AliveOption) *Alive {
	return &Alive{
		option:   option,
		routeMap: sync.Map{},
	}
}

func (a *Alive) findResp(urlpath string) any {
	var urlpaths []string
	urlpaths = append(urlpaths, urlpath)
	if strings.HasSuffix(urlpath, "/") {
		urlpaths = append(urlpaths, urlpath[:len(urlpath)-1])
	} else {
		urlpaths = append(urlpaths, urlpath+"/")
	}

	for _, item := range urlpaths {
		v, ok := a.routeMap.Load(item)
		if ok {
			return v
		}
	}
	return nil
}

func (a *Alive) tryMagic(routePath string, contentType string) string {
	if magic.Hit(routePath) {
		var fileName string
		var p string
		for i := 0; i < magic.RetryCount; i++ {
			fileName = magic.RebuildUrl(routePath, rand.Intn(magic.RetryCount), contentType)
			gologger.Info().Msg(fileName)
			p = filepath.Join(a.option.HomeDir, fileName)
			if utils.IsFileExist(p) {
				return p
			}
		}
	}
	return ""
}

// loadResp is parse ResponseResult by route path/url
func (a *Alive) loadResp(routePath string) *RouteResp {
	fileName := routePath
	if routePath == "/" {
		fileName = "index.html"
	} else if strings.HasSuffix(routePath, "/") {
		fileName = path.Join(routePath, "index.html") // https://github.com/yangyang5214/clone-alive/issues/19
	}

	v := a.findResp(routePath)
	if v == nil {
		gologger.Info().Msgf("routePath <%s> not exist, skip", routePath)
		return nil
	}
	r := v.(*types.ResponseResult)
	p := filepath.Join(a.option.HomeDir, fileName)

	fileSuffix := types.ConvertFileName(r.ResponseContentType)
	if !utils.IsFileExist(p) {
		//http://localhost:8001/SAAS/jersey/manager/api/images/5101/
		// SAAS/jersey/manager/api/images/5101/index.png
		p = filepath.Join(a.option.HomeDir, routePath, "index."+fileSuffix)
	}

	if !utils.IsFileExist(p) {
		p = a.tryMagic(routePath, r.ResponseContentType)
	}

	if !utils.IsFileExist(p) {
		p = p + "." + types.ConvertFileName(r.ResponseContentType) // ?
	}

	if !utils.IsFileExist(p) {
		return nil
	}

	return &RouteResp{
		Resp: r,
		Path: p,
	}
}

func (a *Alive) handleStaticFileRoute() gin.HandlerFunc {
	return func(c *gin.Context) {
		fullPath := strings.Split(c.Request.RequestURI, "?")[0]
		findPath := utils.FindFileByName(a.option.HomeDir, utils.GetSplitLast(fullPath, "/"))
		if findPath == "" {
			c.JSON(http.StatusNotFound, nil)
		} else {
			c.File(findPath)
		}
		return
	}
}

func (a *Alive) handleRoute() gin.HandlerFunc {
	return func(c *gin.Context) {
		fullPath := strings.Split(c.Request.RequestURI, "?")[0]

		//fullPath => /login.action
		//Request.RequestURI => /login.action?language=da_DK
		routeResp := a.loadResp(c.Request.RequestURI)
		if routeResp == nil {
			// http://localhost:8001/?module=captcha&0.06911867290494 验证码
			findPath := a.tryMagic(c.Request.RequestURI, "")
			if findPath != "" {
				c.File(findPath)
				return
			}
			routeResp = a.loadResp(fullPath)
		}

		if routeResp == nil {
			//find by file name
			findPath := utils.FindFileByName(a.option.HomeDir, utils.GetSplitLast(fullPath, "/"))
			if findPath == "" {
				findPath = a.tryMagic(c.Request.RequestURI, "")
			}

			if findPath == "" {
				c.JSON(http.StatusNotFound, nil)
			} else {
				c.File(findPath)
			}
			return
		}
		r := routeResp.Resp
		p := routeResp.Path

		contentType := r.RequestContentType
		contentType = types.ConvertContentType(contentType)
		if contentType == "" {
			contentType = r.ResponseContentType
		}
		c.Header("Content-Type", contentType)

		data, err := utils.ReadFile(p)
		if err != nil {
			c.JSON(http.StatusOK, nil)
			return
		}

		switch contentType {
		case types.ApplicationJson:
			c.Data(r.Status, types.ApplicationJson, data)
		case types.ImagePng, types.ImageJpeg:
			c.Data(r.Status, contentType, data)
		default:
			c.File(p)
		}
	}
}

func (a *Alive) handle(engine *gin.Engine) (err error) {
	f, err := os.Open(a.option.RouteFile)
	defer f.Close()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		var resp *types.ResponseResult
		err = json.Unmarshal([]byte(line), &resp)
		if err != nil {
			gologger.Error().Msg(err.Error())
			continue
		}
		if resp.Error != "" {
			continue
		}

		urlPath := utils.GetUrlPath(resp.Url)

		// - /mas/front/css/index.css
		// - /mas/front//css/index.css
		urlPath = strings.Replace(urlPath, "//", "/", -1)
		_, ok := a.routeMap.Load(urlPath)
		if ok {
			continue
		}

		if types.IsStaticFile(resp.Url) && !magic.Hit(urlPath) {
			engine.Handle(resp.HttpMethod, urlPath, a.handleStaticFileRoute())
		} else {
			// https://stackoverflow.com/questions/32443738/setting-up-route-not-found-in-gin/
			engine.NoRoute(a.handleRoute())
			//engine.Handle(resp.HttpMethod, urlPath, a.handleRoute())
		}

		//https://github.com/yangyang5214/clone-alive/issues/18
		if strings.HasSuffix(urlPath, "woff2") {
			engine.Handle(resp.HttpMethod, urlPath[:len(urlPath)-1], a.handleRoute())
		}
		a.routeMap.Store(urlPath, resp)
	}
	return nil
}

func (a *Alive) check() error {
	if _, err := os.Stat(a.option.HomeDir); err != nil {
		return errors.Wrapf(err, "%s not exist", a.option.HomeDir)
	}
	if _, err := os.Stat(a.option.RouteFile); err != nil {
		return errors.Wrapf(err, "%s not exist", a.option.RouteFile)
	}
	return nil
}

func cleanup() {

}

func (a *Alive) Run() (err error) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(1)
	}()

	err = a.check()
	if err != nil {
		return err
	}
	if !a.option.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	err = a.handle(r)
	if err != nil {
		return err
	}
	server := fmt.Sprintf(":%d", a.option.Port)
	gologger.Info().Msgf("Alive server start with 127.0.0.1%s", server)
	return r.Run(server)
}
