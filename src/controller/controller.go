package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"vicinity-tinymesh-adapter-co2/src/config"
	"vicinity-tinymesh-adapter-co2/src/vicinity"
)

type Server struct {
	config   *config.ServerConfig
	vicinity *vicinity.Client
}

func (server *Server) setupRouter() *gin.Engine {
	r := gin.Default()

	// THING DESCRIPTION
	r.GET("/objects", func(c *gin.Context) {
		c.JSON(http.StatusOK, server.vicinity.TD)
	})

	r.GET("/objects/:oid/properties/:prop", func(c *gin.Context) {
		serveProperties(server, c)
	})

	return r
}

func New(serverConfig *config.ServerConfig, vicinity *vicinity.Client) *Server {
	return &Server{
		vicinity: vicinity,
		config:   serverConfig,
	}
}

// Goroutine
func (server *Server) Listen() {
	router := server.setupRouter()

	err := router.Run(fmt.Sprintf(":%s", server.config.Port))
	if err != nil {
		panic(err)
	}
}
