package redis_rs

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"strconv"
	"strings"
	"time"
)

type RedisRecordStore struct {
	pool      *redis.Pool
	namespace string
}

type RedisRecordStoreOpts struct {
	Database  int
	Namespace string
}

func Create(opts string) (r *RedisRecordStore, err error) {
	options := strings.Split(opts, ",")
	addr := "127.0.0.1:6379"
	db := 0
	ns := "gloon"
	if len(options) >= 1 && options[0] != "" {
		addr = options[0]
	}
	if len(options) >= 2 {
		db, _ = strconv.Atoi(options[1])
	}
	if len(options) >= 3 {
		ns = options[2]
	}
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr, redis.DialDatabase(db))
		},
	}
	r = &RedisRecordStore{pool: pool, namespace: ns}
	return
}

func (r *RedisRecordStore) PutVal(dnsType uint16, key, val string) (err error) {
	conn := r.pool.Get()
	defer conn.Close()
	_, err = conn.Do("SADD", r.keyPath(dnsType, key), val)
	return
}

func (r *RedisRecordStore) Clear() (err error) {
	conn := r.pool.Get()
	defer conn.Close()
	resp, err := conn.Do("KEYS", fmt.Sprintf("/%s/*", r.namespace))
	keys, err := redis.Strings(resp, err)
	if err != nil {
		return
	}
	for _, k := range keys {
		_, err := conn.Do("DEL", k)
		if err != nil {
			return err
		}
	}
	return
}

func (r *RedisRecordStore) GetAll(dnsType uint16, key string) (vals []string, err error) {
	conn := r.pool.Get()
	defer conn.Close()
	vals, err = redis.Strings(conn.Do("SMEMBERS", r.keyPath(dnsType, key)))
	if err == redis.ErrNil {
		err = nil
	}
	return
}

func (r *RedisRecordStore) DelKey(dnsType uint16, key string) (err error) {
	conn := r.pool.Get()
	defer conn.Close()
	_, err = conn.Do("DEL", r.keyPath(dnsType, key))
	if err == redis.ErrNil {
		err = nil
	}
	return
}

func (r *RedisRecordStore) DelVal(dnsType uint16, key, value string) (err error) {
	conn := r.pool.Get()
	defer conn.Close()
	_, err = conn.Do("SREM", r.keyPath(dnsType, key), value)
	if err == redis.ErrNil {
		err = nil
	}
	return
}

func (r *RedisRecordStore) keyPath(dnsType uint16, key string) string {
	return fmt.Sprintf("/%s/%d/%s", r.namespace, dnsType, key)
}
