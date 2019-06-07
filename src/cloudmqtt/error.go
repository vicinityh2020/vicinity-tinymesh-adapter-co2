package cloudmqtt

import "fmt"

type NotCO2SensorError struct {
	sensor *VitirSensorEvent
	msg string
}

func (e NotCO2SensorError) Error() string {
	return fmt.Sprintf("%s: actual dif %d", e.msg, e.sensor.Dif)
}
