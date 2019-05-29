package model

import (
	"fmt"
	"time"
)

type SensorValue struct {
	Now    int
	Hourly int
	Daily  int
}

type Sensor struct {
	Pk          int `storm:"id,increment"` // primary key
	ModelNumber string
	Unit        string
	// Unique id is in format: serialNumber-manufacturer
	UniqueID    string      `storm:"unique"`
	Value       SensorValue `storm:"inline"`
	LastUpdated time.Time   `storm:"index"`
}

func (s *Sensor) GetEid() string {
	return fmt.Sprintf("%s-event", s.UniqueID)
}
