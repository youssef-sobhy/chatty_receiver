package internal

import (
	"context"
	"os"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
)

var (
	redisAddr     = os.Getenv("REDIS_ADDR")
	redisPassword = os.Getenv("REDIS_PASSWORD")
	redisDB       = os.Getenv("REDIS_DB")
	ctx           = context.Background()
)

// CachedNumber func
func CachedNumber(model string, token string, chatNumber int) int {
	// Connect to Redis
	redisDB, _ := strconv.Atoi(redisDB)
	mu := sync.Mutex{}
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	var number int
	var key string
	var query string

	switch model {
	case "Chat":
		key = token
		query = ChatNumberQuery
	case "Message":
		key = token + "-" + strconv.Itoa(chatNumber)
		query = MessageNumberQuery
	}

	mu.Lock()
	lastNumber, err := rdb.Get(ctx, key).Result()
	mu.Unlock()

	if err != nil {
		number = GetNumber(query, token, chatNumber)
	} else {
		lastNumber, _ := strconv.Atoi(lastNumber)
		number = lastNumber + 1
	}

	mu.Lock()
	rdb.Set(ctx, key, number, 0)
	mu.Unlock()
	return number
}
