package record_set

import (
	"fmt"
	"github.com/miekg/dns"
	"log"
	"net"
	"sort"
	"strings"
	"sync"
)

//
// Key/value store used for dns records. keys are unique, but each key has a list of values
type RecordStore interface {
	PutVal(dnsType uint16, key, val string) error        // Put a single key/value type into store.
	GetVal(dnsType uint16, key string) (string, error)   // Gets a single value. Will round robin or randomly select  if multiple values
	GetAll(dnsType uint16, key string) ([]string, error) // Get all key values
	DelKey(dnsType uint16, key string) error             // Deletes key and all values for a
	DelVal(dnsType uint16, key, value string) error      // Deletes a single value from a key. Deletes key ifthere are no more values
	Clear() error                                        // Clear all keys from set
}

type RrIndexes struct {
	sync.Mutex
	indexes map[string]int
}

func (rri *RrIndexes) NextVal(dnsType uint16, key string, vals []string) (val string) {
	vl := len(vals)
	if vl == 0 {
		return
	}
	if vl == 1 {
		val = vals[0]
		return
	}
	sort.Strings(vals)
	rri.Lock()
	defer rri.Unlock()
	kp := fmt.Sprintf("/%d/%s", dnsType, key)
	idx := rri.indexes[kp]
	if idx >= vl {
		idx = 0
	}
	val = vals[idx]
	rri.indexes[kp] = idx + 1
	return
}

func (rri *RrIndexes) Del(dnsType uint16, key string) {
	kp := fmt.Sprintf("/%d/%s", dnsType, key)
	delete(rri.indexes, kp)
}

type RecordSet struct {
	store      RecordStore
	rr_indexes *RrIndexes
}

func Create(store RecordStore) (rs *RecordSet) {
	rs = &RecordSet{store, &RrIndexes{indexes: make(map[string]int)}}
	return
}

func (r *RecordSet) Put(dnsType uint16, host, addr string) {
	log.Printf("Adding/updating  %X %s A %s", dnsType, host, addr)
	err := r.store.PutVal(dnsType, host+".", addr)
	if err != nil {
		log.Printf("Unable to put primary record: %s", err.Error())
		return
	}
	// For A or AAAA records, put in reverse DNS
	if dnsType == dns.TypeA || dnsType == dns.TypeAAAA {
		raddr, _ := ReverseAddr(addr)
		if raddr != "" {
			log.Printf("Adding %s PTR %s", raddr, host)
			err = r.store.PutVal(dns.TypePTR, raddr+".", host)
			if err != nil {
				log.Printf("Error %s addting PTR record %s => %s", raddr, host)
			}
		}
	}
}

func (r *RecordSet) Del(dnsType uint16, host string) {
	log.Printf("Removing %X  %s", dnsType, host)
	addrs, err := r.store.GetAll(dnsType, host+".")
	if err != nil {
		log.Printf("Unable to fetch address for host %s -- %s", host, err.Error())
	}
	for _, addr := range addrs {
		raddr, _ := ReverseAddr(addr)
		err = r.store.DelKey(dns.TypePTR, raddr+".")
		if err != nil {
			log.Printf("Unable to remove PTR record %s -- %s", raddr, err.Error())
		}
	}
	err = r.store.DelKey(dnsType, host+".")
	if err != nil {
		log.Printf("Unable to remove host key %s (%s)", host, err.Error())
	}
	r.rr_indexes.Del(dnsType, host)
}

func (r *RecordSet) DelAddr(dnsType uint16, host, addr string) {
	log.Printf("Removing %X  %s %s", dnsType, host, addr)
	err := r.store.DelVal(dnsType, host+".", addr)
	if err != nil {
		log.Printf("Unable to delete  address %s for host %s -- %s", addr, host, err.Error())
		return
	}
	raddr, _ := ReverseAddr(addr)
	err = r.store.DelKey(dns.TypePTR, raddr+".")
	if err != nil {
		log.Printf("Unable to remove PTR record %s -- %s", raddr, err.Error())
	}
	r.rr_indexes.Del(dnsType, host)
}

func (r *RecordSet) Get(dnsType uint16, host string) (addr string) {
	addrs, err := r.store.GetAll(dnsType, host)
	if err != nil {
		log.Printf("Unable to fetch value: %s", err.Error())
		return
	}
	if len(addrs) == 0 { // Try a wildcard
		parts := strings.SplitN(host, ".", 2)
		if len(parts) == 2 {
			wc := "*." + parts[1]
			addrs, err = r.store.GetAll(dnsType, wc)
			if err != nil {
				log.Printf("Unable to fetch value: %s", err.Error())
				return
			}
		}
	}
	if len(addrs) == 0 { // Try adouble wildcard
		parts := strings.SplitN(host, ".", 3)
		if len(parts) == 3 {
			wc := "*.*." + parts[2]
			addrs, err = r.store.GetAll(dnsType, wc)
			if err != nil {
				log.Printf("Unable to fetch value: %s", err.Error())
				return
			}
		}
	}
	return r.rr_indexes.NextVal(dnsType, host, addrs)
}

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
