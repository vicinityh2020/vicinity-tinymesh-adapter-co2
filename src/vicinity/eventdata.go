package vicinity

import "fmt"

type EventData struct {
	TimeStamp int64
	UniqueID  string
	Value     int
	Unit      string
	ResUnit   string // resUnit
}

func (e *EventData) getEid() string {
	return fmt.Sprintf("%s-event", e.UniqueID)
}