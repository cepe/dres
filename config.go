package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

type Config struct {
	CIDRS         map[string]string         `json:"cidrs"`
	Resolvers     map[string]ResolverConfig `json:"resolvers"`
	Configuration map[string][]string       `json:"configuration"`
}

type ResolverConfig struct {
	Type   string            `json:"type"`
	Socket string            `json:"socket,omitempty"`
	Hosts  map[string]string `json:"hosts,omitempty"`
	Path   string            `json:"path,omitempty"`
}

func LoadConfig() Config {
	configFilePath := flag.String("config", "./config.json", "Path to JSON config file.")
	flag.Parse()

	configFile, err := os.ReadFile(*configFilePath)
	if err != nil {
		log.Fatalf("Unable to read config file. Error: %s", err)
	}

	var config Config
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("Unable to parse configuration file. Error: %s", err)
	}
	return config
}
