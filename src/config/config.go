package config

import (
	"log"
	"os"
)

const (
	defaultMQTTServer = ""
	defaultMQTTPort   = ""
	defaultMQTTTopic  = "#"

	defaultAgentUrl = ""
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
}

type Config struct {
	MQTT     *MQTTConfig
	Vicinity *VicinityConfig
}

// New returns a new Config struct
func New() *Config {
	return &Config{
		MQTT: &MQTTConfig{
			Server:   getEnv("CLOUDMQTT_SERVER", defaultMQTTServer),
			Username: getEnv("CLOUDMQTT_USERNAME", ""),
			Password: getEnv("CLOUDMQTT_PASSWORD", ""),
			Port:     getEnv("CLOUDMQTT_PORT", defaultMQTTPort),
			Topic:    getEnv("CLOUDMQTT_TOPIC", defaultMQTTTopic),
			Secure:   false,
		},
		Vicinity: &VicinityConfig{
			AgentUrl: getEnv("VICINITY_AGENT_URL", defaultAgentUrl),
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
