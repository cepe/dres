package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

type DresConfig struct {
	CIDRS         map[string]string         `json:"cidrs"`
	Resolvers     map[string]ResolverConfig `json:"resolvers"`
	Configuration map[string][]string       `json:"configuration"`
}

type ResolverConfig struct {
	Type   string            `json:"type"`
	Socket string            `json:"socket,omitempty"`
	Hosts  map[string]string `json:"hosts,omitempty"`
}

func LoadConfig() DresConfig {
	configFilePath := flag.String("config", "./config.json", "Path to JSON config file.")
	flag.Parse()

	configFile, err := os.ReadFile(*configFilePath)
	if err != nil {
		log.Fatalf("Unable to read config file. Error: %s", err)
	}

	var config DresConfig
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("Unable to parse configuration file. Error: %s", err)
	}
	return config
}
func LoadResolver(name string, config ResolverConfig) (Resolver, error) {
	if config.Type == "delegating" {
		return DelegatingResolver{
			Name:   name,
			Socket: config.Socket,
		}, nil
	}
	if config.Type == "static" {
		return StaticHostsResolver{
			Name:  name,
			Hosts: config.Hosts,
		}, nil
	}
	errorMessage := fmt.Sprintf("unable to construct resolver of type %s", config.Type)
	return nil, errors.New(errorMessage)
}

func LoadDres(config DresConfig) Dres {
	var resolverByName = make(map[string]Resolver)

	for resolverName, resolverConfig := range config.Resolvers {
		resolver, err := LoadResolver(resolverName, resolverConfig)
		if err != nil {
			log.Fatalf("Error constructing resolver: %s", err)
		} else {
			resolverByName[resolverName] = resolver
			log.Printf("Loaded resolver %s of type %s", resolverName, resolverConfig.Type)
		}
	}

	var networks []Network
	for rangeName, cidr := range config.CIDRS {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Fatalf("Error reading CIDR: %s", err)
		} else {
			network := Network{
				Name: rangeName,
				Net:  *ipNet,
			}
			networks = append(networks, network)
			log.Printf("Loaded network %s: %s", rangeName, cidr)
		}
	}

	var resolvers = make(map[string][]Resolver)
	for rangeName, resolverNames := range config.Configuration {
		log.Printf("Range %s has following resolvers:", rangeName)
		_, ok := resolvers[rangeName]
		if !ok {
			resolvers[rangeName] = make([]Resolver, len(resolverNames))
		}
		for i, resolverName := range resolverNames {
			resolvers[rangeName][i] = resolverByName[resolverName]
			log.Printf(" - %s", resolverName)
		}
	}

	return Dres{
		Networks:  networks,
		Resolvers: resolvers,
	}
}
