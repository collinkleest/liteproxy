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

func GetRequest(key string) (string, error) {
	rdc := GetRedisClient()
	val, err := rdc.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
			panic(err)
	}
	return val, nil
}

func SetRequest(key string, value string) {
	rdc := GetRedisClient()
	err := rdc.Set(ctx, key, value, 0).Err()
	if err != nil {
			panic(err)
	}
}