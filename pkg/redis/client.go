package redis

import redisClient "github.com/go-redis/redis"

type redis struct {
}

func NewRedis() *redis {
	return &redis{}
}

func (r *redis) Redis() *redisClient.Client {
	client := redisClient.NewClient(&redisClient.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	return client
}
