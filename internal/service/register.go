package service

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func NewService(di *do.Injector) *gin.Engine {
	app := gin.New()

	app.Use(gin.Recovery())
	app.Use(gin.Logger())

	{
		handler := NewHelloworldService(di)
		event := app.Group("/hello")
		event.GET("/world", handler.Hello)
	}

	return app
}
