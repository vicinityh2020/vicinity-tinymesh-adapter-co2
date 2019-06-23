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
	"sync"
	"time"
	"vicinity-tinymesh-adapter-co2/src/config"
	"vicinity-tinymesh-adapter-co2/src/model"
)

type EventEmitter struct {
	config     *config.VicinityConfig
	db         *storm.DB
	incoming   chan *EventData
	httpClient *http.Client
	wg         *sync.WaitGroup
	events     []string
}

const (
	timeout = 5 * time.Second
)

func (c *Client) NewEventEmitter(eventCh chan *EventData, wg *sync.WaitGroup) *EventEmitter {
	return &EventEmitter{
		config:   c.config,
		db:       c.db,
		incoming: eventCh,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		wg:     wg,
		events: []string{},
	}
}

func (emitter *EventEmitter) start() {
	var running = true

	for running {
		select {
		case event, ok := <-emitter.incoming:
			if ok {
				if err := emitter.publish(event); err != nil {
					log.Println(err.Error())
				}
			} else {
				running = false
			}
		}
	}

	emitter.wg.Done()
	log.Println("Emitter shut down")
}

// Goroutine
func (emitter *EventEmitter) ListenAndEmit() {

	var sensors []model.Sensor

	if err := emitter.db.All(&sensors); err != nil {
		panic(err)
	}

	// Open event channel for each sensor
	for _, sensor := range sensors {
		if err := emitter.openEventChannel(&sensor); err != nil {
			log.Println("Could not open event channel for", sensor.UniqueID, ":", err.Error())
		}

		emitter.events = append(emitter.events, sensor.GetEid())
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

func (emitter *EventEmitter) send(e *EventData, eid string) error {
	eventPath := fmt.Sprintf("/agent/events/%s", eid)
	uri := emitter.config.AgentUrl + eventPath

	event := map[string]interface{}{
		"value":     e.Value.Now,
		"unit":      e.Unit,
		"timestamp": strconv.FormatInt(e.TimeStamp, 10),
	}

	payload, err := json.Marshal(&event)

	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, uri, bytes.NewBuffer(payload))

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("infrastructure-id", e.UniqueID)
	req.Header.Set("adapter-id", emitter.config.AdapterID)

	req.Close = true

	resp, err := emitter.httpClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	log.Println(string(body))
	return nil
}

func (emitter *EventEmitter) publish(e *EventData) error {
	var eid = e.getEid()

	if !emitter.isEventSupported(eid) {
		return errors.New(fmt.Sprintf("event %s not supported", eid))
	}

	return emitter.send(e, eid)
}
