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
	config     *config.VicinityConfig
	db         *storm.DB
	TD         *gin.H
	eventPipe  chan EventData
}

type EventDescription struct {
	Name    string
	ResUnit string
}

var (
	co2Meta = []Field{
		// Field Value
		{
			Name: "value",
			Schema: Schema{
				Type: "number",
			},
		},

		// Field Unit
		{
			Name: "unit",
			Schema: Schema{
				Type: "string",
			},
		},

		// Field Timestamp
		{
			Name: "timestamp",
			Schema: Schema{
				Type: "string",
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

func (c *Client) makeProperty(description string, pid string, oid string) Property {
	return Property{
		Pid:      pid,
		Monitors: "adapters:CO2Concentration",
		ReadLink: Link{
			Href: fmt.Sprintf("/objects/%s/properties/%s", oid, pid),
			Output: IO{
				Type:        "object",
				Description: description,
				Fields:      co2Meta,
			},
		},
	}
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
	var description = "CO2"

	var events []Event
	var properties []Property

	events = append(events, c.makeEvent(description, sensor.UniqueID))
	properties = append(properties, c.makeProperty(description, "value", sensor.UniqueID))

	return Device{
		Oid:      sensor.UniqueID,
		Name:     fmt.Sprintf("Vitir CO2 Sensor %s", sensor.UniqueID),
		Type:     "core:Device",
		Version:  sensor.ModelNumber,
		Keywords: []string{"co2", "sensor"},

		Properties: properties,
		Actions:    []interface{}{},

		Events: events,
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
