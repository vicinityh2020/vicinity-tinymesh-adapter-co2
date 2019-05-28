package main

import (
	"fmt"
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

func (app *Environment) run() {
	var wg sync.WaitGroup
	defer app.DB.Close()
	defer wg.Wait()

	mqttc := cloudmqtt.New(app.Config.MQTT, app.DB)
	mqttc.Listen()
	defer mqttc.Shutdown()

	v := vicinity.New(app.Config.Vicinity, app.DB)

	emitter := v.NewEventEmitter(mqttc.GetEventChannel(), &wg)
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

func init() {
	// loads values from .app into the system
	if err := godotenv.Load(); err != nil {
		log.Fatalln("No .app file found")
	}

	app.Config = config.New()

	// open bolt db
	db, err := storm.Open("my.db", storm.BoltOptions(0600, &bolt.Options{Timeout: 1 * time.Second}))
	if err != nil {
		log.Fatalln(err.Error())
	}

	app.DB = db
	// uncomment for mock data
	insertMockData(app.DB)
}

func main() {
	// init is invoked before main automatically
	app.run()
}
