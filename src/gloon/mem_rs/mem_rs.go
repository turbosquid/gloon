package mem_rs

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type MemRecordStore struct {
	sync.Mutex
	data map[string][]string
}

func Create() (rs *MemRecordStore) {
	rs = &MemRecordStore{data: make(map[string][]string)}
	rand.Seed(time.Now().UnixNano())
	return
}

func (rs *MemRecordStore) PutVal(dnsType uint16, key, val string) (err error) {
	rs.Lock()
	defer rs.Unlock()
	vals := rs.data[keyPath(dnsType, key)]
	vals = append(vals, val)
	rs.data[keyPath(dnsType, key)] = vals
	return
}

func (rs *MemRecordStore) GetVal(dnsType uint16, key string) (val string, err error) {
	rs.Lock()
	defer rs.Unlock()
	vals := rs.data[keyPath(dnsType, key)]
	if vals == nil || len(vals) == 0 {
		return
	}
	if len(vals) == 1 {
		val = vals[0]
		return
	}

	// Return a random value
	val = vals[rand.Intn(len(vals))]
	return
}

func (rs *MemRecordStore) GetAll(dnsType uint16, key string) (vals []string, err error) {
	rs.Lock()
	defer rs.Unlock()
	vals = rs.data[keyPath(dnsType, key)]
	if vals == nil {
		vals = []string{}
	}
	return
}

func (rs *MemRecordStore) DelKey(dnsType uint16, key string) (err error) {
	rs.Lock()
	defer rs.Unlock()
	delete(rs.data, keyPath(dnsType, key))
	return
}

func (rs *MemRecordStore) DelVal(dnsType uint16, key, val string) (err error) {
	rs.Lock()
	defer rs.Unlock()
	kp := keyPath(dnsType, key)
	vals := rs.data[kp]
	if vals == nil {
		return
	}
	for i, v := range vals {
		if v == val {
			vals = append(vals[:i], vals[i+1:]...)
			rs.data[kp] = vals
			break
		}
	}
	if len(vals) == 0 {
		delete(rs.data, key)
	}
	return
}

func (rs *MemRecordStore) Clear() (err error) {
	rs.Lock()
	defer rs.Unlock()
	rs.data = make(map[string][]string)
	return
}

func keyPath(dnsType uint16, key string) string {
	return fmt.Sprintf("%d/%s", dnsType, key)
}
