package main

import "net"

func GetIP(addr net.Addr) net.IP {
	switch addr := addr.(type) {
	case *net.UDPAddr:
		return addr.IP
	case *net.TCPAddr:
		return addr.IP
	}
	return nil
}
