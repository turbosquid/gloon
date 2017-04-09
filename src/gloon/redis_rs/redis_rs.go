package redis_rs

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
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

func Create(address string, opts *RedisRecordStoreOpts) (r *RedisRecordStore, err error) {
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", address, redis.DialDatabase(opts.Database))
		},
	}
	r = &RedisRecordStore{pool: pool, namespace: "gloon"}
	if opts.Namespace != "" {
		r.namespace = opts.Namespace
	}
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

func (r *RedisRecordStore) GetVal(dnsType uint16, key string) (val string, err error) {
	conn := r.pool.Get()
	defer conn.Close()
	v, err := conn.Do("SRANDMEMBER", r.keyPath(dnsType, key)) // Not RR -- we just pull a random value
	val, err = redis.String(v, err)
	if err == redis.ErrNil {
		err = nil
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
