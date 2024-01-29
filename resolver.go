package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type Resolver interface {
	Handle(msg *dns.Msg) (*dns.Msg, error)
	GetName() string
}

type DelegatingResolver struct {
	Name   string
	Socket string
}

type StaticHostsResolver struct {
	Name  string
	Hosts map[string]string
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
	if config.Type == "hosts-file" {
		started := time.Now()
		hosts, err := ReadHostsMapping(config.Path)
		if err != nil {
			return nil, err
		}
		log.Printf("Resolver %s loaded in %d ms", name, time.Now().Sub(started).Milliseconds())

		return StaticHostsResolver{
			Name:  name,
			Hosts: hosts,
		}, nil
	}
	errorMessage := fmt.Sprintf("unable to construct resolver of type %s", config.Type)
	return nil, errors.New(errorMessage)
}

func ReadHostsMapping(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)

	hostsMapping := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" && !strings.HasPrefix(line, "#") {
			mapping := strings.Split(line, " ")
			hostsMapping[mapping[1]] = mapping[0]
		}
	}

	return hostsMapping, nil
}

func (resolver DelegatingResolver) Handle(msg *dns.Msg) (*dns.Msg, error) {
	return dns.Exchange(msg, resolver.Socket)
}

func (resolver DelegatingResolver) GetName() string {
	return resolver.Name
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
