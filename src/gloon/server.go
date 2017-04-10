package main

import (
	"fmt"
	"github.com/miekg/dns"
	"gloon/mem_rs"
	"gloon/record_set"
	"gloon/redis_rs"
	"log"
	"regexp"
)

type Server struct {
	*dns.Server
	*record_set.RecordSet
	resolver *Resolver
	settings *Settings
}

var split_rex = regexp.MustCompile("[:= ]")

func NewServer(addr string, settings *Settings) (s *Server, err error) {
	s = &Server{}
	s.settings = settings

	// Set up the record store
	var store record_set.RecordStore
	switch settings.Store {
	case "redis":
		store, err = redis_rs.Create(settings.StoreOpts)
		if err != nil {
			log.Fatalf("Unable to create redis record set: %s", err.Error())
		}
	case "memory":
		store = mem_rs.Create()
	default:
		log.Fatalf("Unknown dns record store type %s specified", settings.Store)
	}
	s.RecordSet = record_set.Create(store)
	s.Server = &dns.Server{Addr: addr, Net: "udp"}
	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		s.handleDnsRequest(w, r)
	})
	s.resolver, err = NewResolver(settings)
	for _, v := range settings.Hostnames {
		parts := split_rex.Split(v, -1)
		if len(parts) == 2 {
			s.RecordSet.Put(dns.TypeA, parts[0], parts[1])
		}
	}
	return
}

func (s *Server) handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	defer func() {
		if r := recover(); r != nil {
			handlePanic(r)
		}
	}()
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false
	switch r.Opcode {
	case dns.OpcodeQuery:
		if s.processQuery(m) {
			w.WriteMsg(m)
			return
		}
	}
	if s.settings.DisableForwarding {
		log.Printf("WARNING: name not found and forwarding disabled")
		m.Rcode = dns.RcodeNameError
		w.WriteMsg(m)
		return
	}
	resp, err := s.resolver.Lookup(r)
	if err == nil {
		w.WriteMsg(resp)
	} else {
		log.Printf("Resolver err: %s", err.Error())
	}
}

func (s *Server) processQuery(m *dns.Msg) bool {
	answers := 0

	for _, q := range m.Question {
		var rr dns.RR
		switch q.Qtype {
		case dns.TypeA:
			ip := s.Get(dns.TypeA, q.Name)
			if ip != "" {
				rr, _ = dns.NewRR(fmt.Sprintf("%s %d A %s", q.Name, s.settings.Ttl, ip))
			}
		case dns.TypePTR:
			host := s.Get(dns.TypePTR, q.Name)
			if host != "" {
				rr, _ = dns.NewRR(fmt.Sprintf("%s %d PTR %s", q.Name, s.settings.Ttl, host))
			}
		case dns.TypeAAAA:
			ip := s.Get(dns.TypeAAAA, q.Name)
			if ip != "" {
				rr, _ = dns.NewRR(fmt.Sprintf("%s %d AAAA %s", q.Name, s.settings.Ttl, ip))
			}
		default:
			return false // If we get a question we can't answer, bail
		}
		if rr != nil {
			m.Answer = append(m.Answer, rr)
			answers++
		}
	}
	if answers > 0 {
		return true
	}
	return false
}
