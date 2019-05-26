package vicinity

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asdine/storm"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"vicinity-tinymesh-adapter-co2/src/config"
	"vicinity-tinymesh-adapter-co2/src/model"
)

type EventEmitter struct {
	config     *config.VicinityConfig
	db         *storm.DB
	incoming   chan EventData
	httpClient *http.Client
	events     []string
}

func (c *Client) NewEventEmitter(eventPipe chan EventData) *EventEmitter {
	return &EventEmitter{
		config:     c.config,
		db:         c.db,
		incoming:   eventPipe,
		httpClient: &http.Client{},
		events:     []string{},
	}
}

func (emitter *EventEmitter) start() {
	select {
	case event, ok := <-emitter.incoming:
		if ok {
			fmt.Printf("Value %d was read.\n", event.Value)
			err := emitter.publish(&event)
			if err != nil {
				log.Println(err.Error())
			}
		} else {
			fmt.Println("Channel closed!")
		}
	default:
		fmt.Println("No value ready, moving on.")
	}
}

func (emitter *EventEmitter) ListenAndEmit() {

	var sensors []model.Sensor

	if err := emitter.db.All(&sensors); err != nil {
		panic(err)
	}

	// Open event channel for each sensor
	for _, sensor := range sensors {
		if err := emitter.openEventChannel(&sensor); err != nil {
			log.Println("Could not open event channel for", sensor.UniqueID, " at eid:")
		}
	}

	emitter.start()
}

func (emitter *EventEmitter) isEventSupported(eid string) bool {
	for _, e := range emitter.events {
		if eid == e {
			return true
		}
	}
	return false
}

func (emitter *EventEmitter) openEventChannel(sensor *model.Sensor) error {
	return errors.New("not implemented")
}

func (emitter *EventEmitter) publish(e *EventData) error {
	var eventID = e.getEid()

	if !emitter.isEventSupported(eventID) {
		return errors.New(fmt.Sprintf("event %s not supported", eventID))
	}

	eventPath := fmt.Sprintf("/agent/events/%s", eventID)
	uri := emitter.config.AgentUrl + eventPath

	event := map[string]interface{}{
		"value":     e.Value,
		"unit":      e.Unit,
		"timestamp": strconv.FormatInt(e.TimeStamp, 10),
	}

	payload, err := json.Marshal(&event)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, uri, bytes.NewBuffer(payload))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("infrastructure-id", e.UniqueID)
	req.Header.Set("adapter-id", emitter.config.AdapterID)

	if err != nil {
		return err
	}

	resp, err := emitter.httpClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	// todo: replace with status checks
	log.Println(body)

	return nil
}
