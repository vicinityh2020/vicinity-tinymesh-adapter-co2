package cloudmqtt

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/asdine/storm"
	"github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"vicinity-tinymesh-adapter-co2/src/config"
	"vicinity-tinymesh-adapter-co2/src/vicinity"
)

const (
	quiesce = 250
)

type Client struct {
	config *config.MQTTConfig
	db     *storm.DB
	eventPipe chan vicinity.EventData
}

func (cmqtt *Client) updateDb(sensor *VitirSensor) error {
	// todo: update db
	return nil
}

func (cmqtt *Client) forwardEvent(co2Data *VitirSensor) {
	event := vicinity.EventData{
		TimeStamp: co2Data.TimeStamp,
		UniqueID:  co2Data.UniqueID,
	}

	for _, point := range co2Data.Datapoint {

		var value int
		if err := json.Unmarshal(point.Value, &value); err != nil {
			log.Println(fmt.Sprintf("Could not unmarshal value of a datapoint: %s", err.Error()))
			continue
		}

		event.Value = value
		event.Unit = point.Unit
		event.ResUnit = point.Unit

		cmqtt.eventPipe <- event
	}
}

func (cmqtt *Client) registerCallback() mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {
		// extract the co2 relevant sensor data
		co2Data, err := extractCO2Data(message.Payload())
		if err != nil {
			log.Println(err.Error())
			return
		}

		// forward event to vicinity EventEmitter
		cmqtt.forwardEvent(co2Data)

		// update the local database
		if err := cmqtt.updateDb(co2Data); err != nil {
			log.Println(err.Error())
		}
	}
}

func (cmqtt *Client) buildMQTTConnection() mqtt.Client {
	var onMessageCallback = cmqtt.registerCallback()

	mqtt.ERROR = log.New(os.Stdout, "", 0)
	mqtt.DEBUG = log.New(os.Stdout, "", 0)

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

func New(env *config.MQTTConfig, db *storm.DB) *Client {

	eventChannel := make(chan vicinity.EventData)
	client := &Client{
		config: env,
		db:     db,
		eventPipe: eventChannel,
	}

	return client
}

// Goroutine
func (cmqtt *Client) Listen() {

	mqttClient := cmqtt.buildMQTTConnection()

	defer func() {
		mqttClient.Disconnect(quiesce)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		log.Println("Connected to", cmqtt.config.Server)
	}
	<-c
}

func (cmqtt *Client) GetEventPipe() chan vicinity.EventData {
	return cmqtt.eventPipe
}