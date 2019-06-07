package model

import (
	"fmt"
)

type SensorValue struct {
	Now    int
	Hourly int
	Daily  int
}

type Sensor struct {
	UniqueID     string `storm:"id"`
	ModelNumber  string
	Unit         string
	Value        SensorValue `storm:"inline"`
	LastUpdated  int64       `storm:"index"`
}

func (s *Sensor) GetEid() string {
	return fmt.Sprintf("%s-event", s.UniqueID)
}
