package database

import (
	"fmt"
	"sync"

	"github.com/arizon-dread/plats/internal/config"
	"github.com/gomodule/redigo/redis"
)

type Cache struct {
}

var instance *Cache = nil

var m *sync.Mutex

func New() *Cache {
	for {
		if instance == nil {
			if locked := m.TryLock(); locked {
				instance = new(Cache)
				m.Unlock()
				break
			}
		} else {
			break
		}
	}
	return instance
}

func (c *Cache) Store(key string, value any) error {
	conn, err := conn()
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.Do("SET", key, value)
	return nil
}

func (c *Cache) Get(key string) (string, error) {
	conn, err := conn()
	if err != nil {
		return "", err
	}
	defer conn.Close()
	val, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return "", err
	}
	return val, nil
}

func conn() (redis.Conn, error) {
	conf := config.Load()

	conn, err := redis.Dial(conf.Cache.Proto, fmt.Sprintf("%v:%v", conf.Cache.Url, conf.Cache.Port))
	if err != nil {
		return nil, fmt.Errorf("error getting redis connection %v", err)
	}
	return conn, nil
}
