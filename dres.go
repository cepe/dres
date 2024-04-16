package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sort"
)
import "github.com/miekg/dns"

type Network struct {
	Name string
	Net  net.IPNet
}

type Networks []Network

func (networks Networks) Len() int { return len(networks) }
func (networks Networks) Less(i, j int) bool {
	s1, _ := networks[i].Net.Mask.Size()
	s2, _ := networks[j].Net.Mask.Size()
	return s1 > s2
}
func (networks Networks) Swap(i, j int) { networks[i], networks[j] = networks[j], networks[i] }

type Dres struct {
	Networks  Networks
	Resolvers map[string][]Resolver
}

func Load(config Config) Dres {
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

	var networks Networks
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

	sort.Sort(networks)

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

func (dres Dres) GetResolvers(addr net.Addr) []Resolver {
	networkName, err := dres.GetNetworkName(addr)
	if err != nil {
		log.Printf("Network for address %s not found: %s", addr, err)
		return make([]Resolver, 0)
	}
	return dres.Resolvers[networkName]
}

func (dres Dres) HandleFunc(writer dns.ResponseWriter, msg *dns.Msg) {
	log.Printf("Request from %s", writer.RemoteAddr())
	for _, question := range msg.Question {
		log.Printf("  Question: %d", question.String())
	}

	for _, resolver := range dres.GetResolvers(writer.RemoteAddr()) {
		response, err := resolver.Handle(msg)
		if err != nil {
			log.Printf("    Resolver %s failed to handle query: %s", resolver.GetName(), err)
		} else {
			log.Printf("    Answer from resolver %s", resolver.GetName())
			if writer.WriteMsg(response) != nil {
				log.Printf("    Unable to response. See error %s", err)
			} else {
				return
			}
		}
	}
	log.Printf("    Query from %s not handled", writer.RemoteAddr())
	_ = writer.Close()
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
	dres := Load(LoadConfig())

	server := dns.Server{Addr: ":53", Net: "udp"}

	dns.HandleFunc(".", dres.HandleFunc)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Unable to start dres. See error: %s", err)
	}
}
