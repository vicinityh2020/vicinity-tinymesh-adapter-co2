package model

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type SensorValue struct {
	Now    int
	Hourly int
	Daily  int
}

type Sensor struct {
	UniqueID    string `storm:"id"`
	ModelNumber string
	Unit        string
	Value       SensorValue `storm:"inline"`
	LastUpdated int64       `storm:"index"`
	Latitude    float64
	Longitude   float64
}

func (s *Sensor) GetEid() string {
	return fmt.Sprintf("%s-event", s.UniqueID)
}

func (s *Sensor) GetValue() gin.H {
	return gin.H{
		"value":     s.Value.Now,
		"unit":      s.Unit,
		//"timestamp": s.LastUpdated,
	}
}

func (s *Sensor) GetLatitude() gin.H {
	return gin.H{
		"latitude": s.Latitude,
	}
}

func (s *Sensor) GetLongitude() gin.H {
	return gin.H{
		"longitude": s.Longitude,
	}
}
