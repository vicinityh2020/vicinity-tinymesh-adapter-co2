package vicinity

import (
	"fmt"
	"github.com/asdine/storm"
	"github.com/gin-gonic/gin"
	"log"
	"vicinity-tinymesh-adapter-co2/src/config"
	"vicinity-tinymesh-adapter-co2/src/model"
)

type Client struct {
	config  *config.VicinityConfig
	db      *storm.DB
	TD      *gin.H
	eventCh chan EventData
}

type EventDescription struct {
	Name    string
	ResUnit string
}

const (
	// Fields
	semanticValue       = "core:value"
	semanticTimestamp   = "core:timestamp"
	s4bldgBuildingSpace = "s4bldg:BuildingSpace"

	// Devices
	coreDevice        = "core:Device"
	adaptersCo2Sensor = "adapters:CO2Sensor"

	// Monitors
	adaptersLatitude         = "adapters:GPSLatitude"
	adaptersLongitude        = "adapters:GPSLongitude"
	adaptersCO2Concentration = "adapters:CO2Concentration"
)

var (
	co2Meta = []Field{
		// Field Value
		{
			Name:        "value",
			Description: "co2 reading",
			Predicate:   semanticValue,
			Schema: Schema{
				Type: "integer",
			},
		},

		// Field Unit
		{
			Name:        "unit",
			Description: "co2 measurement unit",
			Schema: Schema{
				Type: "string",
			},
		},

		// Field Timestamp
		{
			Name:        "timestamp",
			Description: "Unix timestamp of time the reading was received",
			Predicate:   semanticTimestamp,
			Schema: Schema{
				Type: "integer",
			},
		},
	}

	latitudeMeta = []Field{
		{
			Name:        "latitude",
			Description: "latitudinal coordinates of the device",
			Schema: Schema{
				Type: "double",
			},
		},
	}

	longitudeMeta = []Field{
		{
			Name:        "longitude",
			Description: "longitudinal coordinates of the device",
			Schema: Schema{
				Type: "double",
			},
		},
	}
)

func New(vicinityConfig *config.VicinityConfig, db *storm.DB) *Client {
	v := &Client{
		config: vicinityConfig,
		db:     db,
	}

	v.makeTD()
	return v
}

func (c *Client) makeProperty(description string, monitors string, pid string, oid string, fields []Field, staticValue interface{}) Property {

	prop := Property{
		Pid:      pid,
		Monitors: monitors,
		ReadLink: Link{
			Href: fmt.Sprintf("/objects/%s/properties/%s", oid, pid),
			Output: IO{
				Type:        "object",
				Description: description,
				Fields:      fields,
			},
		},
	}

	if staticValue != nil {
		prop.ReadLink.StaticValue = map[string]interface{}{
			pid: staticValue,
		}
	}

	return prop
}

func (c *Client) makeEvent(description string, oid string) Event {
	return Event{
		Eid:      fmt.Sprintf("%s-event", oid),
		Monitors: "adapters:CO2Concentration",
		Output: IO{
			Type:        "object",
			Description: description,
			Fields:      co2Meta,
		},
	}
}

func (c *Client) makeDevice(sensor model.Sensor) Device {
	var sensorDescription = "CO2"

	var events []Event
	var properties []Property

	events = append(events, c.makeEvent(sensorDescription, sensor.UniqueID))
	properties = append(properties, c.makeProperty(sensorDescription, adaptersCO2Concentration, "value", sensor.UniqueID, co2Meta, nil))
	properties = append(properties, c.makeProperty("latitudinal coordinates", adaptersLatitude, "latitude", sensor.UniqueID, latitudeMeta, sensor.Latitude))
	properties = append(properties, c.makeProperty("longitudinal coordinates", adaptersLongitude, "longitude", sensor.UniqueID, longitudeMeta, sensor.Longitude))

	return Device{
		Oid:      sensor.UniqueID,
		Name:     fmt.Sprintf("Vitir CO2 Sensor %s", sensor.UniqueID),
		Type:     adaptersCo2Sensor,
		Version:  sensor.ModelNumber,        // only for services?
		Keywords: []string{"co2", "sensor"}, // only for services?

		Properties: properties,
		Actions:    []interface{}{},
		Events:     events,
		LocatedIn: []Location{
			{LocationType: s4bldgBuildingSpace, LocationId: "https://www.cwi.no", Label: "CWi Moss"},
		},
	}
}

func (c *Client) makeTD() {
	if c.TD != nil {
		return
	}

	var sensors []model.Sensor
	var devices []Device

	if err := c.db.All(&sensors); err != nil {
		log.Fatalln("could not fetch all sensors from storm DB")
	}

	for _, sensor := range sensors {
		devices = append(devices, c.makeDevice(sensor))
	}

	c.TD = &gin.H{
		"adapter-id":         c.config.AdapterID,
		"thing-descriptions": devices,
	}
}

func (c *Client) GetSensor(oid string) (*model.Sensor, error) {
	var sensor model.Sensor
	err := c.db.One("UniqueID", oid, &sensor)
	return &sensor, err
}
