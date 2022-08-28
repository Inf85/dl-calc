package redis

import (
	"crypto/tls"
	redisClient "github.com/go-redis/redis/v8"
)

type redis struct {
}

func NewRedis() *redis {
	return &redis{}
}

func (r *redis) Redis() *redisClient.ClusterClient {
	client := redisClient.NewClusterClient(&redisClient.ClusterOptions{
		Addrs:     []string{"myredis-001.hubals.0001.use1.cache.amazonaws.com:6379"},
		TLSConfig: &tls.Config{},
	})

	return client
}
