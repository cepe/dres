package main

import "log"
import "github.com/miekg/dns"

func main() {
	server := dns.Server{
		Addr: ":53",
		Net:  "udp",
	}

	dns.HandleFunc(".", handleRequest)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Unable to start dres. See error: %s", err)
	}
}

func handleRequest(writer dns.ResponseWriter, msg *dns.Msg) {
	response, err := dns.Exchange(msg, "1.1.1.1:53")
	if err != nil {
		log.Fatalf("Failed to delegate query. See error %s", err)
	}

	if writer.WriteMsg(response) != nil {
		log.Fatalf("Unable to response. See error %s", err)
	}
}
