package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"window-sliding/ratelimiter"

	"github.com/redis/go-redis/v9"
)

type RateLimiter interface {
	Allow(key string) bool
}

func RateLimitMiddleware(limiter *ratelimiter.RedisSlidingWindow, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := extractClientKey(r)
		allowed, err := limiter.Allow(r.Context(), key)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if !allowed {
			w.Header().Set("Rety-After", "60")
			w.Header().Set("X-RateLimit", "100")
			http.Error(w, "Rate limit exceeded. Please slow down here!", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func extractClientKey(r *http.Request) string {
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return "apikey:" + apiKey
	}

	ip := r.RemoteAddr

	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = strings.Split(forwarded, ",")[0]
		ip = strings.TrimSpace(ip)
	}

	if colonIdx := strings.LastIndex(ip, ":"); colonIdx != -1 {
		if strings.Count(ip, ":") == 1 {
			ip = ip[:colonIdx]
		}
	}

	return "ip:" + ip
}

func handleRequest(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "Request processed successfully at %v\n", time.Now())
}

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	duration := 5 * time.Second

	limiter := ratelimiter.NewRedisSlidingWindow(client, 5, duration)

	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	// Wrap with rate limiting
	http.Handle("/", RateLimitMiddleware(limiter, apiHandler))

	fmt.Println("Server starting on :8081")
	fmt.Println("------------------------------------->")
	http.ListenAndServe(":8081", nil)
}
