package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

var serveProperties = func(server *Server, c *gin.Context) {
	oid, exists := c.Params.Get("oid")
	if !exists {
		c.JSON(http.StatusBadRequest, nil)
	}

	prop, exists := c.Params.Get("prop")
	if !exists {
		c.JSON(http.StatusBadRequest, nil)
	}

	sensor, err := server.vicinity.GetSensor(oid)

	if err != nil {
		c.JSON(http.StatusNotFound, nil)
	}

	switch prop {
	case "value":
		c.JSON(http.StatusOK, gin.H{
			"value":     sensor.Value.Instant,
			"unit":      sensor.Unit,
			"timestamp": sensor.LastUpdated,
		})
	default:
		c.JSON(http.StatusNotFound, nil)
	}
}
