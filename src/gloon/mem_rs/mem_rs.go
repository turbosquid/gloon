package mem_rs

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type RecData map[string]map[string]bool

type MemRecordStore struct {
	sync.RWMutex
	data RecData
}

func Create() (rs *MemRecordStore) {
	rs = &MemRecordStore{data: make(RecData)}
	rand.Seed(time.Now().UnixNano())
	return
}

func (rs *MemRecordStore) PutVal(dnsType uint16, key, val string) (err error) {
	rs.Lock()
	defer rs.Unlock()
	vals := rs.data[keyPath(dnsType, key)]
	if vals == nil {
		vals = make(map[string]bool)
	}
	vals[val] = true
	rs.data[keyPath(dnsType, key)] = vals
	return
}

func (rs *MemRecordStore) GetAll(dnsType uint16, key string) (vals []string, err error) {
	rs.RLock()
	defer rs.RUnlock()
	valmap := rs.data[keyPath(dnsType, key)]
	if valmap == nil {
		vals = []string{}
	}
	vals = getKeysFromMap(valmap)
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
	delete(vals, val)
	rs.data[kp] = vals
	if len(vals) == 0 {
		delete(rs.data, key)
	}
	return
}

func (rs *MemRecordStore) Clear() (err error) {
	rs.Lock()
	defer rs.Unlock()
	rs.data = make(RecData)
	return
}

func keyPath(dnsType uint16, key string) string {
	return fmt.Sprintf("%d/%s", dnsType, key)
}

func getKeysFromMap(valmap map[string]bool) (keys []string) {
	keys = make([]string, len(valmap))
	i := 0
	for k, _ := range valmap {
		keys[i] = k
		i++
	}
	return
}
