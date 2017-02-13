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
		r.recs[dnsType] = make(IpMap)
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
	delete(r.recs[dnsType], host+".")
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
