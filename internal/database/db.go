package database

import (
	"fmt"

	"github.com/arizon-dread/plats/internal/config"
	"github.com/gomodule/redigo/redis"
)

type Db struct {
}

func (c *Db) Store(key string, value any) error {
	conn, err := conn()
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.Do("SET", key, value)
	return nil
}

func (c *Db) Get(key string) *string {
	conn, err := conn()
	if err != nil {
		return nil
	}
	val, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return nil
	}
	return &val
}

func conn() (redis.Conn, error) {
	conf := config.Config{}
	conf.Load()

	conn, err := redis.Dial(conf.Cache.Proto, conf.Cache.Url)
	if err != nil {
		return nil, fmt.Errorf("error getting redis connection %v", err)
	}
	return conn, nil
}
