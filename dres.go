package main

import (
	"errors"
	"fmt"
	"log"
	"net"
)
import "github.com/miekg/dns"

type Network struct {
	Name string
	Net  net.IPNet
}
type Dres struct {
	Networks  []Network
	Resolvers map[string][]Resolver
}

func (dres Dres) HandleFunc(writer dns.ResponseWriter, msg *dns.Msg) {
	log.Printf("Request from %s", writer.RemoteAddr())
	for _, question := range msg.Question {
		log.Printf("  Question %s", question.Name)
	}

	for _, resolver := range dres.GetResolvers(writer.RemoteAddr()) {
		response, err := resolver.Handle(msg)
		if err != nil {
			log.Printf("Resolver %s failed to handle query: %s", resolver.GetName(), err)
		} else {
			log.Printf("Answer from resolver %s", resolver.GetName())
			if writer.WriteMsg(response) != nil {
				log.Printf("Unable to response. See error %s", err)
			} else {
				return
			}
		}
	}
	log.Printf("Query from %s not handled", writer.RemoteAddr())
	_ = writer.Close()
}

func (dres Dres) GetResolvers(addr net.Addr) []Resolver {
	networkName, err := dres.GetNetworkName(addr)
	if err != nil {
		log.Printf("Network for address %s not found: %s", addr, err)
		return make([]Resolver, 0)
	}
	return dres.Resolvers[networkName]
}

func GetIP(addr net.Addr) net.IP {
	switch addr := addr.(type) {
	case *net.UDPAddr:
		return addr.IP
	case *net.TCPAddr:
		return addr.IP
	}
	return nil
}
func (dres Dres) GetNetworkName(addr net.Addr) (string, error) {
	for _, network := range dres.Networks {
		if network.Net.Contains(GetIP(addr)) {
			return network.Name, nil
		}
	}
	errorMessage := fmt.Sprintf("unable to find network for %s", addr.String())
	return "", errors.New(errorMessage)
}

func main() {
	config := LoadConfig()
	dres := LoadDres(config)

	server := dns.Server{
		Addr: ":53",
		Net:  "udp",
	}

	dns.HandleFunc(".", dres.HandleFunc)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Unable to start dres. See error: %s", err)
	}
}
