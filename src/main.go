package main

import (
	"fmt"
	"github.com/asdine/storm"
	"github.com/joho/godotenv"
	bolt "go.etcd.io/bbolt"
	"log"
	"os"
	"os/signal"
	"path"
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
	Config  *config.Config
	DB      *storm.DB
	LogPath string
}

var app Environment

func (app *Environment) syncDb() {

	var sensors = []model.Sensor{
		{
			UniqueID:    "LAS00016225",
			ModelNumber: "LAN-WMBUS-E-CO2",
			Unit:        "ppm",
			Value: model.SensorValue{
				Now:    0,
				Hourly: 0,
				Daily:  0,
			},
			LastUpdated: time.Now().Unix(),
			Latitude:    59.4407535,
			Longitude:   10.6667105,
		},
		{
			UniqueID:    "LAS00016222",
			ModelNumber: "LAN-WMBUS-E-CO2",
			Unit:        "ppm",
			Value: model.SensorValue{
				Now:    0,
				Hourly: 0,
				Daily:  0,
			},
			LastUpdated: time.Now().Unix(),
			Latitude:    59.4407617,
			Longitude:   10.6667319,
		},
	}

	for _, s := range sensors {
		var d model.Sensor
		if err := app.DB.One("UniqueID", s.UniqueID, &d); err == storm.ErrNotFound {
			if err := app.DB.Save(&s); err != nil {
				if err != storm.ErrAlreadyExists {
					log.Fatalln(err.Error())
				}
			}
		}
	}
}

func (app *Environment) init() {
	// loads values from .env into the system

	app.LogPath = path.Join(".", "logs")
	if err := os.MkdirAll(app.LogPath, os.ModePerm); err != nil {
		log.Fatal("could not create path:", app.LogPath)
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

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

	// Logger
	mainLogger, err := os.OpenFile(path.Join(app.LogPath, fmt.Sprintf("adapter-%s.log", time.Now().Format("2006-01-02"))), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Fatal("Could not create mainLogger logfile:", err.Error())
	}
	defer mainLogger.Close()

	mqttLogger, err := os.OpenFile(path.Join(app.LogPath, fmt.Sprintf("mqtt-%s.log", time.Now().Format("2006-01-02"))), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Could not create MQTT trace logfile:", err.Error())
	}
	defer mqttLogger.Close()

	ginLogger, err := os.OpenFile(path.Join(app.LogPath, fmt.Sprintf("gin-%s.log", time.Now().Format("2006-01-02"))), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Could not create GIN trace logfile:", err.Error())
	}
	defer ginLogger.Close()

	log.SetOutput(mainLogger)

	defer app.DB.Close()
	defer wg.Wait()

	// MQTT
	mqttc := cloudmqtt.New(app.Config.MQTT, app.DB, log.New(mqttLogger, "", log.Ldate|log.Ltime))
	mqttc.Listen()
	defer mqttc.Shutdown()

	// VICINITY
	v := vicinity.New(app.Config.Vicinity, app.DB)

	// Event Emitter
	emitter := v.NewEventEmitter(mqttc.GetEventChannel(), &wg)

	wg.Add(1)
	go emitter.ListenAndEmit()

	// Controller
	server := controller.New(app.Config.Server, v, ginLogger)
	go server.Listen()
	defer server.Shutdown()

	// INT handler
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
	app.run()
}
