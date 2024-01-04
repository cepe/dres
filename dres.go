package main

import "log"
import "github.com/miekg/dns"

func main() {
	server := dns.Server{
		Addr: ":53",
		Net:  "udp",
	}

	err := server.ListenAndServe()

	if err != nil {
		log.Fatalf("Unable to start dres. See error: %s", err)
	}
}
