package cloudmqtt

import (
	"encoding/json"
	"log"
	"strings"
	"time"
)

type Predicate func(DataPoint) bool

var byUnit Predicate = func(point DataPoint) bool {
	return strings.Contains(point.Unit, "ppm")
}

const (
	co2SensorDif = 42

	resUnitNow    = "ppm"
	resUnitHourly = "Hour"
	resUnitDaily  = "24hour"
)

type DataPoint struct {
	Type      string          `json:"type"`
	Name      string          `json:"name"`
	Size      int             `json:"size"`
	DataType  string          `json:"dataType"`
	Unit      string          `json:"unit"`
	Res       float64         `json:"res"`
	ResUnit   string          `json:"resUnit"`
	ValueType string          `json:"valueType"`
	Value     json.RawMessage `json:"value"`
	Scale     float64         `json:"scale"`
	Min       string          `json:"min"`
	Max       string          `json:"max"`
	Low       string          `json:"low"`
	High      string          `json:"high"`
}

type VitirSensorEvent struct {
	DsType       string      `json:"dsType"`
	MrfCuID      string      `json:"mrfCuId"`
	TimeStamp    int64       `json:"timeStamp"`
	DateTime     time.Time   `json:"dateTime"`
	SerialNo     string      `json:"serialNo"`
	Manufacturer string      `json:"manufacturer"`
	ModelNo      string      `json:"modelNo"`
	BattLvl      int         `json:"battLvl"`
	BridgeID     string      `json:"bridgeId"`
	Rssi         int         `json:"rssi"`
	HopCnt       int         `json:"hopCnt"`
	LatCnt       int         `json:"latCnt"`
	DpCnt        int         `json:"dpCnt"`
	Datapoint    []DataPoint `json:"datapoint"`
	Vif          int         `json:"vif"`
	Dif          int         `json:"dif"` // 42 == co2
	RssiWmbus    int         `json:"rssiWmbus"`
	UniqueID     string      `json:"uniqueId"`
}

func filter(dataPoints []DataPoint, predicate Predicate) (ret []DataPoint) {
	for _, point := range dataPoints {
		if predicate(point) {
			ret = append(ret, point)
		}
	}
	return ret
}

func extractCO2Data(payload []byte) (*VitirSensorEvent, error) {
	var sensor VitirSensorEvent

	err := json.Unmarshal(payload, &sensor)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	if sensor.Dif != co2SensorDif {
		return nil, &NotCO2SensorError{msg: "dif does not indicate a co2 sensor", sensor: &sensor}
	}

	log.Println("co2 sensor data arrived")

	sensor.Datapoint = filter(sensor.Datapoint, byUnit)

	return &sensor, nil
}
