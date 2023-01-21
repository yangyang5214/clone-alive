package alive

import "C"
import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/projectdiscovery/gologger"
	"github.com/yangyang5214/clone-alive/pkg/types"
	"github.com/yangyang5214/clone-alive/pkg/utils"
	"net/http"
	"os"
	"path/filepath"
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
		}

		p := filepath.Join(a.option.HomeDir, fileName)
		data, err := utils.ReadFile(p)
		if err != nil {
			c.JSON(http.StatusOK, "")
		}

		v, ok := a.routeMap.Load(fullPath)
		if !ok {
			c.JSON(http.StatusOK, "")
		}
		r := v.(*types.ResponseResult)

		contentType := r.RequestContentType
		if contentType == "" {
			contentType = r.ResponseContentType
		}
		c.Header("Content-Type", types.ConvertContentType(contentType))

		switch contentType {
		case types.ApplicationJson:
			c.JSON(r.Status, data)
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
		if urlPath == "" {
			urlPath = "/"
		}
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
	return r.Run(fmt.Sprintf(":%d", a.option.Port))
}
