package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)
import "github.com/miekg/dns"

type DresConfig struct {
	Resolvers map[string]ResolverConfig `json:"resolvers"`
}

type ResolverConfig struct {
	Type   string `json:"type"`
	Socket string `json:"socket,omitempty"`
}

func main() {
	configFilePath := flag.String("config", "/etc/dres/config.json", "Path to JSON config file.")
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

	server := dns.Server{
		Addr: ":53",
		Net:  "udp",
	}

	dns.HandleFunc(".", buildHandleFunc(config))

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Unable to start dres. See error: %s", err)
	}
}

func buildHandleFunc(config DresConfig) func(dns.ResponseWriter, *dns.Msg) {

	var socket string

	for resolverName, resolver := range config.Resolvers {
		if resolver.Type == "delegating" {
			socket = resolver.Socket
			log.Printf("Found resolver named `%s` - delegating resolver, using socket %s", resolverName, socket)
			break
		}
	}

	if socket == "" {
		log.Fatalf("Unable to find delegating resolver.")
	}

	return func(writer dns.ResponseWriter, msg *dns.Msg) {
		response, err := dns.Exchange(msg, socket)
		log.Printf("Request from %s", writer.RemoteAddr())
		for idx, question := range msg.Question {
			log.Printf("  question %d: %s", idx, question.Name)
		}
		if err != nil {
			log.Printf("Failed to delegate query. See error %s", err)
			_ = writer.Close()
			return
		}

		if writer.WriteMsg(response) != nil {
			log.Printf("Unable to response. See error %s", err)
		}
	}
}
