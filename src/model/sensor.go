package model

import "time"

type Reading struct {
	Instant int
	Hourly  int
	Daily   int
}

type Sensor struct {
	Pk           int `storm:"id,increment"` // primary key
	SerialNumber string
	ModelNumber  string
	Unit         string
	UniqueID     string `storm:"unique"`
	Value        Reading `storm:"inline"`
	LastUpdated  time.Time `storm:"index"`
}
