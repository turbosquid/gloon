package record_set

import (
	"github.com/miekg/dns"
	"gloon/mem_rs"
	"testing"
)

func TestAll(t *testing.T) {
	r := mem_rs.Create()
	r.Clear()
	rs := Create(r)
	rs.Put(dns.TypeA, "test.example.com", "1.2.3.4")
	addr := rs.Get(dns.TypeA, "test.example.com.")
	if addr != "1.2.3.4" {
		t.Errorf("rs.Get() unexepected value: %s -- expected 1.2.3.4", addr)
	}
	host := rs.Get(dns.TypePTR, "4.3.2.1.in-addr.arpa.")
	if host != "test.example.com" {
		t.Errorf("rs.Get() unexepected value: %s -- expected test.example.com", host)
	}
	rs.Put(dns.TypeA, "*.example.com", "3.4.5.6")
	addr = rs.Get(dns.TypeA, "test.example.com.")
	if addr != "1.2.3.4" {
		t.Errorf("rs.Get() unexepected value: %s -- expected 1.2.3.4", addr)
	}
	addr = rs.Get(dns.TypeA, "foo.example.com.")
	if addr != "3.4.5.6" {
		t.Errorf("rs.Get() unexepected value: %s -- expected 3.4.5.6", addr)
	}
	rs.Put(dns.TypeA, "*.*.example.com", "10.11.12.13")
	addr = rs.Get(dns.TypeA, "test.example.com.")
	if addr != "1.2.3.4" {
		t.Errorf("rs.Get() unexepected value: %s -- expected 1.2.3.4", addr)
	}
	addr = rs.Get(dns.TypeA, "baz.test.example.com.")
	if addr != "10.11.12.13" {
		t.Errorf("rs.Get() unexepected value: %s -- expected 10.11.12.13", addr)
	}

	// Round robin testing
	rs.Put(dns.TypeA, "test.example.com", "1.2.3.4")
	rs.Put(dns.TypeA, "test.example.com", "1.2.3.5")
	rs.Put(dns.TypeA, "test.example.com", "1.2.3.6")

	if addr = rs.Get(dns.TypeA, "test.example.com."); addr != "1.2.3.4" {
		t.Errorf("Unexpected value: %s", addr)
	}
	if addr = rs.Get(dns.TypeA, "test.example.com."); addr != "1.2.3.5" {
		t.Errorf("Unexpected value: %s", addr)
	}
	if addr = rs.Get(dns.TypeA, "test.example.com."); addr != "1.2.3.6" {
		t.Errorf("Unexpected value: %s", addr)
	}
	if addr = rs.Get(dns.TypeA, "test.example.com."); addr != "1.2.3.4" {
		t.Errorf("Unexpected value: %s", addr)
	}

	// Falls back to wildcard after deleted
	rs.Del(dns.TypeA, "test.example.com")
	if addr = rs.Get(dns.TypeA, "test.example.com."); addr != "3.4.5.6" {
		t.Errorf("rs.Get() unexepected value: %s -- expected 3.4.5.6", addr)
	}
	if addr = rs.Get(dns.TypePTR, "4.3.2.1.in-addr.arpa"); addr != "" {
		t.Errorf("Got non empty value: %s", addr)
	}
	if addr = rs.Get(dns.TypePTR, "5.3.2.1.in-addr.arpa"); addr != "" {
		t.Errorf("Got non empty value: %s", addr)
	}
	rs.Del(dns.TypeA, "test.example.com")
}
