package cloudmqtt

import (
	"crypto/tls"
	"fmt"
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
	client mqtt.Client
}

func registerCallback(e chan vicinity.EventData) mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {
		co2Data, err := extractCO2Data(message.Payload())

		if err != nil {
			log.Println(err.Error())
			return
		}

		event := vicinity.EventData{
			TimeStamp: co2Data.TimeStamp,
			UniqueID: co2Data.UniqueID,
		}

		for _, point := range co2Data.Datapoint {
			event.Value = int(point.Value)
			event.Unit = point.Unit
			event.ResUnit = point.Unit

			e <- event
		}
	}
}

func buildMQTTConnection(env *config.MQTTConfig, eventChannel chan vicinity.EventData) mqtt.Client {
	var onMessageCallback = registerCallback(eventChannel)

	mqtt.ERROR = log.New(os.Stdout, "", 0)
	mqtt.DEBUG = log.New(os.Stdout, "", 0)

	var scheme string
	if env.Secure {
		scheme = "ssl"
	} else {
		scheme = "tcp"
	}

	hostname, _ := os.Hostname()

	server := fmt.Sprintf("%s://%s:%s", scheme, env.Server, env.Port)
	opts := mqtt.NewClientOptions().AddBroker(server)

	opts.SetUsername(env.Username)
	opts.SetPassword(env.Password)
	opts.SetClientID(hostname + strconv.Itoa(time.Now().Second()))

	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	opts.SetTLSConfig(tlsConfig)

	opts.OnConnect = func(c mqtt.Client) {
		if token := c.Subscribe(env.Topic, byte(0), onMessageCallback); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	return mqtt.NewClient(opts)
}

func New(env *config.MQTTConfig) (*Client, chan vicinity.EventData) {

	eventChannel := make(chan vicinity.EventData)
	client := &Client{
		config: env,
		client: buildMQTTConnection(env, eventChannel),
	}

	return client, eventChannel
}

func (cli *Client) Listen() {
	defer func() {
		cli.client.Disconnect(quiesce)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	if token := cli.client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		log.Println("Connected to", cli.config.Server)
	}
	<-c
}
