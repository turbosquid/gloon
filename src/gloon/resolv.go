// local resolver. We forward to this when we don't have an answer
package main

import (
	"fmt"
	"github.com/miekg/dns"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const RESOLV_CONF = "/etc/resolv.conf"

type Resolver struct {
	*dns.ClientConfig
	*dns.Client
	qualifiedServers []string
	settings         *Settings
}

func NewResolver(settings *Settings) (r *Resolver, err error) {
	r = &Resolver{}
	path := settings.ResolvFile
	if path == "" {
		path = RESOLV_CONF
	}
	r.settings = settings
	r.ClientConfig, err = dns.ClientConfigFromFile(path)
	if err != nil {
		return
	}
	localIps := getLocalIps()
	if localIps == nil {
		return nil, fmt.Errorf("Unable to enumerate local IP addresses.")
	}
	// Get fully qualified server names from resolveconf
	for _, server := range r.Servers {
		parts := strings.Split(server, "#")
		host := net.JoinHostPort(parts[0], r.Port)
		if len(parts) == 2 {
			host = net.JoinHostPort(parts[0], parts[1])
		} else {
			host = net.JoinHostPort(parts[0], "53")
		}
		if !isSelf(settings.ResolverAddr, host, localIps) {
			r.qualifiedServers = append(r.qualifiedServers, host)
			log.Printf("Added forwarder: %s", host)
		}
	}
	r.Client = &dns.Client{ReadTimeout: time.Duration(settings.ResolverTimeout) * time.Second, WriteTimeout: time.Duration(settings.ResolverTimeout) * time.Second}
	return
}

func isSelf(myaddr, rhost string, localIps []string) bool {
	my_host, my_port, _ := net.SplitHostPort(myaddr)
	res_host, res_port, _ := net.SplitHostPort(rhost)
	if my_port != res_port { // Our ports are different, so not a match
		return false
	}
	if my_host != "" && my_host != "0.0.0.0" {
		if my_host == res_host { // We are listening on a specific address, so see if we match resolver host
			return true
		} else {
			return false
		}
	}
	// We're listenng on all addresses, so see if the resolver host ip matches anything locqal
	for _, lip := range localIps {
		if lip == res_host {
			return true
		}
	}

	return false
}

func getLocalIps() (results []string) {
	ifaces, err := net.Interfaces()
	// handle err
	if err != nil {
		log.Printf("Unable to enumerate interfaces: %s", err.Error())
		return nil
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Printf("Unable to enumerate addresses: %s", err.Error())
			return nil
		}
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// process IP address
			results = append(results, ip.String())
		}
	}
	return
}

func (r *Resolver) Lookup(req *dns.Msg) (msg *dns.Msg, err error) {
	ch := make(chan *dns.Msg, 1)
	var wg sync.WaitGroup
	for _, ns := range r.qualifiedServers {
		wg.Add(1)
		go r.lookup(req, ns, &wg, ch)
	}
	wg.Wait()
	select {
	case rsp := <-ch:
		return rsp, nil
	default:
		return nil, fmt.Errorf("Query failed for nameservers")
	}
	return
}

func (r *Resolver) lookup(req *dns.Msg, nameserver string, wg *sync.WaitGroup, c chan *dns.Msg) {
	defer wg.Done()
	qname := req.Question[0].Name
	qtype := req.Question[0].Qtype

	rsp, rtt, err := r.Exchange(req, nameserver)
	if err != nil {
		log.Printf("Resolver error on %s (%s) -- %s", nameserver, qname, err.Error())
		return
	}
	if rsp != nil && rsp.Rcode != dns.RcodeSuccess {
		log.Printf("%s (%s) query failed: %v", qname, nameserver, rsp.Rcode)
		if rsp.Rcode == dns.RcodeServerFailure {
			return // Only bail if the server fails
		}
	} else {
		if r.settings.Debug {
			log.Printf("%s (%d) resolved by %s rtt: %d", qname, qtype, nameserver, rtt)
		}
	}
	select {
	case c <- rsp:
	default:
	}
}
