package main

import (
	"fmt"
	"github.com/miekg/dns"
	"log"
)

type Server struct {
	*dns.Server
	*Records
	resolver *Resolver
	settings *Settings
}

func NewServer(addr string, settings *Settings) (s *Server, err error) {
	s = &Server{}
	s.Records = NewRecords()
	s.settings = settings
	s.Server = &dns.Server{Addr: addr, Net: "udp"}
	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		s.handleDnsRequest(w, r)
	})
	s.resolver, err = NewResolver(settings)
	s.Records.PutPairs(settings.Hostnames)
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
				rr, _ = dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
			}
		case dns.TypePTR:
			host := s.Get(dns.TypePTR, q.Name)
			if host != "" {
				rr, _ = dns.NewRR(fmt.Sprintf("%s PTR %s", q.Name, host))
			}
		case dns.TypeAAAA: // For now, if we have an A record, for a name, always NXDOMAIN for corresponding AAAA records
			ip := s.Get(dns.TypeAAAA, q.Name)
			if ip != "" {
				rr, _ = dns.NewRR(fmt.Sprintf("%s AAAA %s", q.Name, ip))
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
