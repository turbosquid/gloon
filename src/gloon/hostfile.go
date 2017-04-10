package main

import (
	"github.com/miekg/dns"
	"github.com/rjeczalik/notify"
	. "gloon/record_set"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var ipv4_regexp = regexp.MustCompile("\\d+\\.\\d+\\.\\d+\\.\\d+")

type HostPair struct {
	host, addr string
}

type Hostfile struct {
	hosts          map[HostPair]bool
	fn             string
	recs           *RecordSet
	reloadInterval int
}

func NewHostfile(fn string, recs *RecordSet, reloadInterval int) (hf *Hostfile) {
	hf = &Hostfile{make(map[HostPair]bool), fn, recs, reloadInterval}
	return
}

func (hf *Hostfile) Run() {
	hfn_abs, err := filepath.Abs(hf.fn)
	if err != nil {
		log.Printf("Unable to get absolute path of %s", err.Error())
		return
	}
	err = hf.loadHosts()
	if err != nil {
		log.Printf("Unable to load hosts file: %s", err.Error())
		return
	}
	// Monitor and load
	if hf.reloadInterval > 0 {
		for {
			time.Sleep(time.Duration(hf.reloadInterval) * time.Second)
			hf.loadHosts() // Ignore errors
		}
	} else {
		c := make(chan notify.EventInfo, 1)
		if err = notify.Watch(filepath.Dir(hf.fn), c, notify.Write); err != nil {
			log.Printf("WARNING: hostfile notifications could not be set up: %s", err.Error())
			return
		}
		defer func() {
			notify.Stop(c)
			close(c)
		}()
		for evt := range c {
			if evt.Path() == hfn_abs {
				log.Printf("Reloading modified hostfile: %s", hf.fn)
				hf.loadHosts() // Ignoring errors
			}
		}
	}
}

func (hf *Hostfile) loadHosts() (err error) {
	hosts := make(map[HostPair]bool)
	hm, err := parseHosts(hf.fn)
	if err != nil {
		return err
	}
	for ip, hostnames := range hm {
		if !ipv4_regexp.MatchString(ip) { // For now we don't do v6
			continue
		}
		for _, hn := range hostnames {
			hp := HostPair{hn, ip}
			if !hf.hosts[hp] { // Dont incur the log cost
				hf.recs.Put(dns.TypeA, hn, ip)
			}
			hosts[hp] = true
		}
	}
	// Remove hosts not in new file
	for hp, _ := range hf.hosts {
		if !hosts[hp] {
			hf.recs.DelAddr(dns.TypeA, hp.host, hp.addr)
		}
	}
	hf.hosts = hosts
	return
}

func parseHosts(fn string) (hm map[string][]string, err error) {
	hm = map[string][]string{}
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		return
	}
	content := string(b)
	for _, line := range strings.Split(content, "\n") {
		line = strings.Replace(strings.Trim(line, "  \t"), "\t", " ", -1)
		if i := strings.Index(line, "#"); i != -1 {
			line = line[0:i]
		}
		if len(line) == 0 {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 && len(parts[0]) > 0 {
			addr := parts[0]
			if names := strings.Fields(parts[1]); len(names) > 0 {
				if _, ok := hm[addr]; ok {
					hm[addr] = append(hm[addr], names...)
				} else {
					hm[addr] = names
				}
			}
		}
	}
	return
}
