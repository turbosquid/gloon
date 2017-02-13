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
}

func NewResolver(path string) (r *Resolver, err error) {
	r = &Resolver{}
	if path == "" {
		path = RESOLV_CONF
	}
	r.ClientConfig, err = dns.ClientConfigFromFile(path)
	if err != nil {
		return
	}
	// Get fully qualified server names from resolveconf
	for _, server := range r.Servers {
		parts := strings.Split(server, "#")
		host := net.JoinHostPort(parts[0], r.Port)
		if len(parts) == 2 {
			host = net.JoinHostPort(parts[0], parts[1])
		}
		r.qualifiedServers = append(r.qualifiedServers, host)
	}
	r.Client = &dns.Client{ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second}
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
		log.Printf("%s resolved by %s rtt: %d", qname, nameserver, rtt)
	}
	select {
	case c <- rsp:
	default:
	}
}
