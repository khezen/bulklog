package redisc

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

// Config - redis config
type Config struct {
	Endpoint string `yaml:"endpoint"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// Connector manage connection to redis
type Connector interface {
	Open() (redis.Conn, error)
}

// Connector manage connection to redis
type connector struct {
	redisEndpoint string
	redisOptions  []redis.DialOption
}

// New redis connector
func New(cfg *Config) Connector {
	return &connector{
		redisEndpoint: cfg.Endpoint,
		redisOptions: []redis.DialOption{
			redis.DialDatabase(cfg.DB),
			redis.DialPassword(cfg.Password),
		},
	}
}

// Open a connection to redis
func (c *connector) Open() (redis.Conn, error) {
	conn, err := redis.Dial(
		"tcp", c.redisEndpoint,
		c.redisOptions...,
	)
	if err != nil {
		return nil, fmt.Errorf("redis.Dial.%s", err)
	}
	return conn, nil
}
