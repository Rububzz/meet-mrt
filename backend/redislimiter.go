package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisLimiter struct {
	client *redis.Client
	limit  int
}

func newRedisLimiter(limit int) *RedisLimiter {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	log.Printf("Redis connected ✓")

	return &RedisLimiter{client: client, limit: limit}
}

func (rl *RedisLimiter) currentKey() string {
	now := time.Now()
	return "search_count:" + strconv.Itoa(now.Year()) + ":" + strconv.Itoa(int(now.Month()))
}

func (rl *RedisLimiter) allow() bool {
	key := rl.currentKey()

	count, err := rl.client.Incr(ctx, key).Result()
	if err != nil {
		log.Printf("Redis error: %v", err)
		return false
	}

	if count == 1 {
		rl.client.Expire(ctx, key, 35*24*time.Hour)
	}

	if count > int64(rl.limit) {
		rl.client.Decr(ctx, key)
		return false
	}

	return true
}

func (rl *RedisLimiter) remaining() int {
	key := rl.currentKey()
	count, err := rl.client.Get(ctx, key).Int()
	if err != nil {
		return rl.limit
	}
	remaining := rl.limit - count
	if remaining < 0 {
		return 0
	}
	return remaining
}
