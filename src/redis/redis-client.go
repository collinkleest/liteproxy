package redis

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient       *redis.Client
	ctx               = context.Background()
	defaultExpiration = 15 * time.Minute
	redisProdUrl      = "liteproxy-redis:6379"
	redisLocalUrl     = "localhost:6379"
)

func getRedisUrl() string {
	ginMode := gin.Mode()
	if ginMode == gin.DebugMode {
		return redisLocalUrl
	}
	return redisProdUrl
}

func InitRedis() {
	redisUrl := getRedisUrl()
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: "", // No password set
		DB:       0,  // Use default DB
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
	err := rdc.Set(ctx, key, value, defaultExpiration).Err()
	if err != nil {
		panic(err)
	}
}

func SetRequestWithExpiration(key string, value string, expiration time.Duration) {
	rdc := GetRedisClient()
	err := rdc.Set(ctx, key, value, expiration).Err()
	if err != nil {
		panic(err)
	}
}
