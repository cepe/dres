package main

import (
	"errors"
	"fmt"
	"log"
	"net"
)
import "github.com/miekg/dns"

type DelegatingResolver struct {
	Name   string
	Socket string
}

func (resolver DelegatingResolver) Handle(msg *dns.Msg) (*dns.Msg, error) {
	return dns.Exchange(msg, resolver.Socket)
}

func (resolver DelegatingResolver) GetName() string {
	return resolver.Name
}

type StaticHostsResolver struct {
	Name  string
	Hosts map[string]string
}

func (resolver StaticHostsResolver) Handle(msg *dns.Msg) (*dns.Msg, error) {
	if len(msg.Question) > 1 {
		return nil, errors.New("unable to handle more than one question")
	}
	question := msg.Question[0]
	if question.Qtype != dns.TypeA {
		return nil, errors.New("unable to handle question other than A")
	}

	hostName := question.Name[:len(question.Name)-1]
	ip, ok := resolver.Hosts[hostName]
	if !ok {
		errorMessage := fmt.Sprintf("static mapping for %s not found", hostName)
		return nil, errors.New(errorMessage)
	}
	msg.Authoritative = true
	dom := question.Name
	msg.Answer = make([]dns.RR, 1)
	msg.Answer[0] = &dns.A{
		Hdr: dns.RR_Header{Name: dom, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0},
		A:   net.ParseIP(ip),
	}
	return msg, nil
}

func (resolver StaticHostsResolver) GetName() string {
	return resolver.Name
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

type Resolver interface {
	Handle(msg *dns.Msg) (*dns.Msg, error)
	GetName() string
}

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
