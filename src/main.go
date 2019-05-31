package main

import (
	"github.com/asdine/storm"
	"github.com/joho/godotenv"
	bolt "go.etcd.io/bbolt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"vicinity-tinymesh-adapter-co2/src/cloudmqtt"
	"vicinity-tinymesh-adapter-co2/src/config"
	"vicinity-tinymesh-adapter-co2/src/controller"
	"vicinity-tinymesh-adapter-co2/src/model"
	"vicinity-tinymesh-adapter-co2/src/vicinity"
)

type Environment struct {
	Config *config.Config
	DB     *storm.DB
}

var app Environment

func (app *Environment) syncDb() {

	var sensors = []model.Sensor{
		{
			ModelNumber: "LAN-WMBUS-E-CO2",
			Unit:        "ppm",
			// Unique id is SerialNo-Manufacturer
			UniqueID: "LAS00016225",
			Value: model.SensorValue{
				Now:    0,
				Hourly: 0,
				Daily:  0,
			},
			LastUpdated: time.Now(),
		},
		{
			ModelNumber: "LAN-WMBUS-E-CO2",
			Unit:        "ppm",
			// Unique id is SerialNo-Manufacturer
			UniqueID: "LAS00016222",
			Value: model.SensorValue{
				Now:    0,
				Hourly: 0,
				Daily:  0,
			},
			LastUpdated: time.Now(),
		},
	}

	for _, s := range sensors{
		if err := app.DB.Save(&s); err != nil {
			if err != storm.ErrAlreadyExists {
				log.Fatalln(err.Error())
			}
		}
	}
}

func (app *Environment) init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Fatalln("No .env file found")
	}

	app.Config = config.New()

	// open bolt db
	db, err := storm.Open(".db", storm.BoltOptions(0600, &bolt.Options{Timeout: 1 * time.Second}))
	if err != nil {
		log.Fatalln(err.Error())
	}

	app.DB = db
}

func (app *Environment) run() {
	var wg sync.WaitGroup
	defer app.DB.Close()
	defer wg.Wait()

	mqttc := cloudmqtt.New(app.Config.MQTT, app.DB)
	mqttc.Listen()
	defer mqttc.Shutdown()

	v := vicinity.New(app.Config.Vicinity, app.DB)

	emitter := v.NewEventEmitter(mqttc.GetEventChannel(), &wg)

	wg.Add(1)
	go emitter.ListenAndEmit()

	server := controller.New(app.Config.Server, v)
	go server.Listen()
	defer server.Shutdown()

	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Println("Adapter shutting down...")
}

// init is invoked before main automatically
func init() {
	app.init()
	app.syncDb()
}

func main() {
	// todo: proper logging
	// todo: fix TD with object event values and object property values\
	// todo: make json config for supported devices
	app.run()
}
