package main

import (
	"github.com/miekg/dns"
	"log"
	"regexp"
	"strings"
	"sync"
)

var split_rex = regexp.MustCompile("[:= ]")

type IpMap map[string]string
type AnswerMap map[uint16]IpMap

type Records struct {
	sync.Mutex
	recs AnswerMap
}

func NewRecords() (r *Records) {
	r = &Records{}
	r.recs = make(AnswerMap)
	r.recs[dns.TypeA] = make(IpMap)
	r.recs[dns.TypeAAAA] = make(IpMap)
	r.recs[dns.TypePTR] = make(IpMap)
	return
}

func (r *Records) PutPairs(pairs []string) {
	for _, v := range pairs {
		parts := split_rex.Split(v, -1)
		if len(parts) == 2 {
			r.Put(dns.TypeA, parts[0], parts[1])
		}
	}
}

func (r *Records) Put(dnsType uint16, host, addr string) {
	log.Printf("Adding/updating  %X %s A %s", dnsType, host, addr)
	r.Lock()
	defer r.Unlock()
	if r.recs[dnsType] == nil {
		return
	}
	// For A or AAAA records, put in reverse DNS
	if dnsType == dns.TypeA || dnsType == dns.TypeAAAA {
		raddr, _ := ReverseAddr(addr)
		if raddr != "" {
			log.Printf("Adding %s PTR %s", raddr, host)
			r.recs[dns.TypePTR][raddr+"."] = host
		}
	}
	r.recs[dnsType][host+"."] = addr
}

func (r *Records) Del(dnsType uint16, host string) {
	log.Printf("Removing %X  %s", dnsType, host)
	r.Lock()
	if r.recs[dnsType] == nil {
		return
	}
	defer r.Unlock()
	addr := r.recs[dnsType][host+"."]
	delete(r.recs[dnsType], host+".")
	if addr != "" {
		raddr, _ := ReverseAddr(addr)
		delete(r.recs[dns.TypePTR], raddr+".")
	}
}

func (r *Records) Get(dnsType uint16, host string) string {
	r.Lock()
	defer r.Unlock()
	if r.recs[dnsType] == nil {
		return ""
	}
	addr := r.recs[dnsType][host]
	if addr == "" { // Try a wildcard
		parts := strings.SplitN(host, ".", 2)
		if len(parts) == 2 {
			wc := "*." + parts[1]
			addr = r.recs[dnsType][wc]

		}
	}
	return addr
}
