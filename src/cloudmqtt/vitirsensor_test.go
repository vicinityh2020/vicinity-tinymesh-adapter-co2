package cloudmqtt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getMockData(mockfile string) []byte {
	dir, err := os.Getwd()
	check(err)
	p := path.Join(dir, "test", mockfile)
	dummy, err := ioutil.ReadFile(p)
	check(err)

	return dummy
}

func TestExtractCO2(t *testing.T) {
	// todo: remake with respect to object event

	dummy := getMockData("roomsensor.json")
	fasit := []interface{}{482, 474, 469}

	sensor, err := extractCO2Data(dummy)
	if err != nil {
		t.FailNow()
	}

	// Asserts
	if sensor.Dif != co2SensorDif {
		t.FailNow()
	}

	for k, v := range sensor.Datapoint {

		var actual int
		if err := json.Unmarshal(v.Value, &actual); err != nil {
			fmt.Println("could not unmarshal v.Value")
			t.FailNow()
		}

		if actual != fasit[k] {
			t.FailNow()
		}
	}
}