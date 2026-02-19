package ratelimiter

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// We are now impleementing a rate limter algorithm + (redis)
// in order to be efficient for distributed system.

type RedisSlidingWindow struct {
	client     *redis.Client
	limit      int
	windowSize time.Duration
	keyPrefix  string
}

func NewRedisSlidingWindow(client *redis.Client, limit int, windowSize time.Duration) *RedisSlidingWindow {
	return &RedisSlidingWindow{
		client:     client,
		limit:      limit,
		windowSize: windowSize,
		keyPrefix:  "ratelimit",
	}
}

func floatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func (r *RedisSlidingWindow) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now()
	windowStart := now.Add(-r.windowSize)
	redisKey := r.keyPrefix + key
	// pipelining makes it super fast (we are able to send multiple comands through a single request to the redis server)
	pipe := r.client.Pipeline()

	pipe.ZRemRangeByScore(ctx, redisKey, "0", floatToString(float64(windowStart.UnixMicro())))

	countCmd := pipe.ZCard(ctx, redisKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count := countCmd.Val()
	if count >= int64(r.limit) {
		return false, nil
	}

	member := float64(now.UnixMicro())
	err = r.client.ZAdd(ctx, redisKey, redis.Z{
		Score:  member,
		Member: member,
	}).Err()
	if err != nil {
		return false, err
	}

	r.client.Expire(ctx, redisKey, r.windowSize*2)
	return true, nil
}
