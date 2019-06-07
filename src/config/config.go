package config

import (
	"log"
	"os"
)

const (
	mqttServer = "localhost"
	mqttPort   = "1663"
	mqttTopic  = "#"

	vicinityAgentUrl = "http://localhost:9997"
	vicinityAdapterID = "15dfe786-272c-4239-a852-be6b7605661f"

	serverPort = "8080"
)

type MQTTConfig struct {
	Server   string
	Username string
	Password string
	Port     string
	Topic    string
	Secure   bool
}

type VicinityConfig struct {
	AgentUrl string
	AdapterID string
}

type ServerConfig struct {
	Port string
}

type Config struct {
	MQTT     *MQTTConfig
	Vicinity *VicinityConfig
	Server   *ServerConfig
}

// New returns a new Config struct
func New() *Config {
	return &Config{
		MQTT: &MQTTConfig{
			Server:   getEnv("CLOUDMQTT_SERVER", mqttServer),
			Username: getEnv("CLOUDMQTT_USERNAME", ""),
			Password: getEnv("CLOUDMQTT_PASSWORD", ""),
			Port:     getEnv("CLOUDMQTT_PORT", mqttPort),
			Topic:    getEnv("CLOUDMQTT_TOPIC", mqttTopic),
			Secure:   false,
		},
		Vicinity: &VicinityConfig{
			AgentUrl: getEnv("VICINITY_AGENT_URL", vicinityAgentUrl),
			AdapterID: getEnv("VICINITY_ADAPTER_ID", vicinityAdapterID),
		},
		Server: &ServerConfig{
			Port: getEnv("SERVER_PORT", serverPort),
		},
	}
}

// Simple helper function to read an environment or return a default value
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	if isEmpty(defaultVal) {
		log.Printf("environment variable %v is empty\n", key)
		os.Exit(0)
	}

	return defaultVal
}

func isEmpty(val string) bool {
	return val == ""
}
