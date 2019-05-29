package vicinity

import "fmt"

type EventValue struct {
	Now    int
	Hourly int
	Daily  int
}

type EventData struct {
	TimeStamp int64
	UniqueID  string
	Value     EventValue
	Unit      string
}

func (e *EventData) getEid() string {
	return fmt.Sprintf("%s-event", e.UniqueID)
}
