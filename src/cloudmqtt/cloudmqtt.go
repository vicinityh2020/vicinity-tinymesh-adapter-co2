package cloudmqtt

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/asdine/storm"
	"github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"strconv"
	"time"
	"vicinity-tinymesh-adapter-co2/src/config"
	"vicinity-tinymesh-adapter-co2/src/model"
	"vicinity-tinymesh-adapter-co2/src/vicinity"
)

const (
	quiesce = 250

	resUnitNow    = "ppm"
	resUnitHourly = "Hour"
	resUnitDaily  = "24hour"
)

type Client struct {
	config          *config.MQTTConfig
	db              *storm.DB
	eventCh         chan *vicinity.EventData
	client          mqtt.Client
	traceLogger     *log.Logger

	// hackathon
	hackathonLogger *log.Logger
	lastTick time.Time
}

func (cmqtt *Client) updateDb(e *vicinity.EventData) error {
	s := model.Sensor{
		UniqueID:    e.UniqueID,
		LastUpdated: e.TimeStamp,
		Value:       model.SensorValue(e.Value),
	}
	return cmqtt.db.Update(&s)
}

func (cmqtt *Client) registerCallback() mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {
		// extract the co2 relevant sensor data
		co2Data, err := extractCO2Data(message.Payload())
		if err != nil {
			return
		}

		var event = translateEventData(co2Data)

		// forward event to vicinity EventEmitter
		cmqtt.eventCh <- event

		// Hackathon
		cmqtt.writeFile(event)

		// update the local database
		if err := cmqtt.updateDb(event); err != nil {
			log.Println(err.Error())
		}
	}
}

func (cmqtt *Client) buildMQTTConnection() mqtt.Client {
	var onMessageCallback = cmqtt.registerCallback()

	mqtt.ERROR = cmqtt.traceLogger
	//mqtt.DEBUG = cmqtt.traceLogger

	var scheme string
	if cmqtt.config.Secure {
		scheme = "ssl"
	} else {
		scheme = "tcp"
	}

	hostname, _ := os.Hostname()

	server := fmt.Sprintf("%s://%s:%s", scheme, cmqtt.config.Server, cmqtt.config.Port)
	opts := mqtt.NewClientOptions().AddBroker(server)

	opts.SetUsername(cmqtt.config.Username)
	opts.SetPassword(cmqtt.config.Password)
	opts.SetClientID(hostname + strconv.Itoa(time.Now().Second()))

	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	opts.SetTLSConfig(tlsConfig)

	opts.OnConnect = func(c mqtt.Client) {
		if token := c.Subscribe(cmqtt.config.Topic, byte(0), onMessageCallback); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	return mqtt.NewClient(opts)
}

func New(env *config.MQTTConfig, db *storm.DB, logger *log.Logger, hackathon *log.Logger) *Client {

	eventChannel := make(chan *vicinity.EventData)
	client := &Client{
		config:  env,
		db:      db,
		eventCh: eventChannel,
		traceLogger: logger,
		hackathonLogger: hackathon,
		lastTick: time.Now(),
	}

	return client
}

// Goroutine
func (cmqtt *Client) Listen() {

	cmqtt.client = cmqtt.buildMQTTConnection()

	if token := cmqtt.client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		log.Println("Connected to", cmqtt.config.Server)
	}

}

func (cmqtt *Client) Shutdown() {
	cmqtt.client.Disconnect(quiesce)
	close(cmqtt.eventCh)
	log.Println("MQTT client shut down")
}

func (cmqtt *Client) GetEventChannel() chan *vicinity.EventData {
	return cmqtt.eventCh
}

func (cmqtt *Client) writeFile(data *vicinity.EventData) {

	now := time.Now()
	if now.Sub(cmqtt.lastTick) >= (1 * time.Hour) {
		cmqtt.traceLogger.Println("Value written for hackathon")
		cmqtt.hackathonLogger.Println(data.Value.Now)
	}

	cmqtt.lastTick = now
}

func translateEventData(co2Data *VitirSensorEvent) *vicinity.EventData {
	e := vicinity.EventData{
		TimeStamp: co2Data.TimeStamp,
		UniqueID:  co2Data.UniqueID,
	}

	for i, point := range co2Data.Datapoint {

		var value int
		if err := json.Unmarshal(point.Value, &value); err != nil {
			log.Println(fmt.Sprintf("Could not unmarshal value of a datapoint: %s", err.Error()))
			continue
		}

		switch point.ResUnit {
		case resUnitNow:
			e.Value.Now = value
			e.Unit = co2Data.Datapoint[i].Unit
			break
		case resUnitHourly:
			e.Value.Hourly = value
			break
		case resUnitDaily:
			e.Value.Daily = value
			break
		default:
			log.Print(co2Data.UniqueID, "contains an unknown resUnit value:", point.ResUnit)
			break
		}
	}

	return &e
}
