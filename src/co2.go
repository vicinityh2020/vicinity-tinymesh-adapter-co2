package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"vicinity-tinymesh-adapter-co2/src/config"
	cloudmqtt "vicinity-tinymesh-adapter-co2/src/cloudmqtt"
)

// init is invoked before main
func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
		os.Exit(0)
	}
}

func main() {
	env := config.New()
	log.Print(env.MQTT.Server)

	mqttc := cloudmqtt.New(env.MQTT)
	mqttc.Listen()
}
