package main

import (
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
	"github.com/miekg/dns"
	. "gloon/record_set"
	"log"
	"net/http"
	"strings"
	"time"
)

var DnsTypes = map[string]uint16{"A": dns.TypeA}

func PanicHandler(w http.ResponseWriter, r *http.Request, p interface{}) {
	handlePanic(p)
	Json(w, "Internal Error", 500)
}

func LogMiddleWare(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	next(rw, r)
	dur := time.Since(start)
	res := rw.(negroni.ResponseWriter)
	log.Printf("Completed: %s %s -- %d (%v)", r.Method, r.URL.Path, res.Status(), dur)
}

func Json(w http.ResponseWriter, text string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, text)
}

func ApiPutHost(w http.ResponseWriter, r *http.Request, ps httprouter.Params, recs *RecordSet) {
	dnsType := strings.ToUpper(ps.ByName("type"))
	host := ps.ByName("host")
	addr := ps.ByName("ip")
	dt, ok := DnsTypes[dnsType]
	if !ok {
		Json(w, "Address type not found", 404)
		return
	}
	recs.Put(dt, host, addr)
	Json(w, "ok", 200)
}

func ApiDelHost(w http.ResponseWriter, r *http.Request, ps httprouter.Params, recs *RecordSet) {
	dnsType := strings.ToUpper(ps.ByName("type"))
	host := ps.ByName("host")
	dt, ok := DnsTypes[dnsType]
	if !ok {
		Json(w, "Address type not found", 404)
		return
	}
	recs.Del(dt, host)
	Json(w, "ok", 200)
}

func ApiDelHostAddr(w http.ResponseWriter, r *http.Request, ps httprouter.Params, recs *RecordSet) {
	dnsType := strings.ToUpper(ps.ByName("type"))
	host := ps.ByName("host")
	addr := ps.ByName("addr")
	dt, ok := DnsTypes[dnsType]
	if !ok {
		Json(w, "Address type not found", 404)
		return
	}
	recs.DelAddr(dt, host, addr)
	Json(w, "ok", 200)
}

func RunApiServer(settings *Settings, recs *RecordSet) {
	router := httprouter.New()
	router.PanicHandler = PanicHandler
	router.PUT("/records/:type/:host/:ip", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ApiPutHost(w, r, ps, recs)
	})
	router.DELETE("/records/:type/:host", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ApiDelHost(w, r, ps, recs)
	})
	router.DELETE("/records/:type/:host/:addr", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		ApiDelHostAddr(w, r, ps, recs)
	})
	n := negroni.New()
	n.Use(negroni.HandlerFunc(LogMiddleWare))
	n.UseHandler(router)
	log.Fatal(http.ListenAndServe(settings.ApiAddr, n))
}
