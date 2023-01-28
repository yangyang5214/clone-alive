package alive

import "C"
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
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type Alive struct {
	option   types.AliveOption
	routeMap sync.Map
}

func New(option types.AliveOption) *Alive {
	return &Alive{
		option:   option,
		routeMap: sync.Map{},
	}
}

func (a *Alive) handleRoute() gin.HandlerFunc {
	return func(c *gin.Context) {
		fullPath := c.FullPath()
		fileName := fullPath
		if fullPath == "/" {
			fileName = "index.html"
		} else if strings.HasSuffix(fullPath, "/") {
			fileName = path.Join(fullPath, "index.html")
		}

		p := filepath.Join(a.option.HomeDir, fileName)

		v, ok := a.routeMap.Load(fullPath)
		if !ok {
			c.JSON(http.StatusOK, "")
		}
		r := v.(*types.ResponseResult)

		contentType := r.RequestContentType
		if contentType == "" || contentType == "<nil>" {
			contentType = r.ResponseContentType
		}
		contentType = types.ConvertContentType(contentType)
		c.Header("Content-Type", contentType)

		if magic.Hit(fullPath, contentType) {
			fileName = magic.RebuildUrl(fullPath, rand.Intn(magic.RetryCount), contentType)
			p = filepath.Join(a.option.HomeDir, fileName)
		}

		data, err := utils.ReadFile(p)
		if err != nil {
			c.JSON(http.StatusOK, "")
		}

		switch contentType {
		case types.ApplicationJson:
			c.JSON(r.Status, data)
		case types.ImagePng:
		case types.ImageJpeg:
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
		_, ok := a.routeMap.Load(urlPath)
		if ok {
			continue
		}
		engine.Handle(resp.HttpMethod, urlPath, a.handleRoute())
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

func (a *Alive) Run() (err error) {
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
