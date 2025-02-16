package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)


var (
	redisClient *redis.Client
	ctx = context.Background()
)

func InitRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:	  "liteproxy-redis:6379",
		Password: "", // No password set
		DB:		  0,  // Use default DB
		Protocol: 2,  // Connection protocol
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
}

func GetRedisClient() *redis.Client {
	if redisClient == nil {
		InitRedis()
	}
	return redisClient
}

