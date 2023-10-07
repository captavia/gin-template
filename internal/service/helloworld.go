package service

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/sirupsen/logrus"
)

type Helloworld struct {
	log *logrus.Entry
}

func NewHelloworldService(di *do.Injector) *Helloworld {
	hw := new(Helloworld)
	logger := logrus.New()
	logger.SetFormatter(do.MustInvoke[logrus.Formatter](di))
	hw.log = logger.WithField("Hello", "")
	return hw
}

func (h *Helloworld) Hello(c *gin.Context) {
	h.log.Info(c.GetQuery("ciallo"))
	c.JSON(200, gin.H{"hello": "world"})
	return
}
