package main

import (
	"fmt"
	"github.com/asdine/storm"
	"github.com/joho/godotenv"
	bolt "go.etcd.io/bbolt"
	"log"
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

func insertMockData(db *storm.DB) {

	var mocks = []model.Sensor{
		{
			SerialNumber: "123",
			ModelNumber:  "A",
			UniqueID:     "A-123",
			Unit:         "ppm",
			Value: model.Reading{
				Instant: 50,
				Hourly:  100,
				Daily:   150,
			},
		},
		{
			SerialNumber: "456",
			ModelNumber:  "B",
			UniqueID:     "B-456",
			Unit:         "ppm",
			Value: model.Reading{
				Instant: 60,
				Hourly:  110,
				Daily:   160,
			},
		},
		{
			SerialNumber: "789",
			ModelNumber:  "C",
			UniqueID:     "C-789",
			Unit:         "ppm",
			Value: model.Reading{
				Instant: 70,
				Hourly:  120,
				Daily:   170,
			},
		},
	}

	for _, m := range mocks {

		if err := db.Save(&m); err != nil {
			log.Println(fmt.Sprintf("#%s %s", m.UniqueID, err.Error()))
		}
	}
}

// init is invoked before main
func init() {
	// loads values from .app into the system
	if err := godotenv.Load(); err != nil {
		log.Fatalln("No .app file found")
	}

	// open bolt db
	db, err := storm.Open("my.db", storm.BoltOptions(0600, &bolt.Options{Timeout: 1 * time.Second}))
	if err != nil {
		log.Fatalln(err.Error())
	}

	app.DB = db
	// uncomment for mock data
	insertMockData(app.DB)

	app.Config = config.New()
}

func main() {
	defer app.DB.Close()

	mqttc, eventPipe := cloudmqtt.New(app.Config.MQTT)
	go mqttc.Listen()

	v := vicinity.New(app.Config.Vicinity, app.DB)

	emitter := v.NewEventEmitter(eventPipe)
	go emitter.ListenAndEmit()

	server := controller.New(app.Config.Server, v)
	server.Listen()
}
