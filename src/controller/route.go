package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (server *Server) ServeProperties(c *gin.Context) {
	oid, exists := c.Params.Get("oid")
	if !exists {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	prop, exists := c.Params.Get("prop")
	if !exists {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	sensor, err := server.vicinity.GetSensor(oid)

	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	switch prop {
	case "value":
		c.JSON(http.StatusOK, gin.H{
			"value":     sensor.Value.Instant,
			"unit":      sensor.Unit,
			"timestamp": sensor.LastUpdated,
		})
		break
	default:
		c.AbortWithStatus(http.StatusNotFound)
		break
	}
}
