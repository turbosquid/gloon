package main

import (
	"fmt"
	"net"
	"strings"
)

// Taken somewhat from stdlib dnsclient.go
func ReverseAddr(addr string) (arpa string, err error) {
	ip := net.ParseIP(addr)
	if ip == nil {
		return "", fmt.Errorf("Unrecognized address: %s", addr)
	}
	if ip.To4() != nil {
		arpa = fmt.Sprintf("%d.%d.%d.%d.in-addr.arpa", ip[15], ip[14], ip[13], ip[12])
		return
	}
	// Must be IPv6
	var parts []string
	for i := len(ip) - 1; i >= 0; i-- {
		v := byte(ip[i])
		str := fmt.Sprintf("%x.%x.", v&0xf, v>>4)
		parts = append(parts, str)
	}
	// Append "ip6.arpa." and return (buf already has the final .)
	parts = append(parts, "ip6.arpa.")
	return strings.Join(parts, ""), nil
}
