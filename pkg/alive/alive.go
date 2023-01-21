package alive

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/yangyang5214/clone-alive/pkg/types"
	"net/http"
	"os"
)

type Alive struct {
	option types.AliveOption
}

func New(option types.AliveOption) *Alive {
	return &Alive{
		option: option,
	}
}

func (a *Alive) handle(r *gin.Engine) {
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
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
	a.handle(r)
	return r.Run(fmt.Sprintf(":%d", a.option.Port))
}
