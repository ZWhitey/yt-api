package model

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// InitRedis 初始化 Redis 連接
func InitRedis() {
	// 從環境變數獲取 Redis URL，如果沒有設置則使用默認值
	redisURL := os.Getenv("REDIS_URL")

	// 使用 ParseURL 來解析 redis://host 格式的 URL
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL %s: %v", redisURL, err)
	}

	RedisClient = redis.NewClient(opt)

	// 測試連接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("Failed to connect to Redis at %s: %v", redisURL, err)
	} else {
		log.Printf("Successfully connected to Redis at %s", redisURL)
	}
}

// CloseRedis 關閉 Redis 連接
func CloseRedis() {
	if RedisClient != nil {
		RedisClient.Close()
	}
}
