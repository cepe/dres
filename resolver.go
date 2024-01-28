package main

import (
	"errors"
	"fmt"
	"github.com/miekg/dns"
	"net"
)

type Resolver interface {
	Handle(msg *dns.Msg) (*dns.Msg, error)
	GetName() string
}
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
