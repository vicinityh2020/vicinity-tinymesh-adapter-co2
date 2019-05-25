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
)

const (
	quiesce = 250
)

var knt int

var onMessageCallback mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("MSG: %s\n", msg.Payload())
	text := fmt.Sprintf("this is result msg #%d!", knt)
	knt++
	token := client.Publish("nn/result", 0, false, text)
	token.Wait()
}

type Client struct {
	config *config.MQTTConfig
	client mqtt.Client
}

func buildMQTTConnection(env *config.MQTTConfig) mqtt.Client {

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

func New(env *config.MQTTConfig) *Client {
	return &Client{
		config: env,
		client: buildMQTTConnection(env),
	}
}

func (cli *Client) Listen() {
	defer func() {
		cli.client.Disconnect(quiesce)
	}()

	knt = 0
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	if token := cli.client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		log.Println("Connected to", cli.config.Server)
	}
	<-c
}
