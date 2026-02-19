package main

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"bucket-alg/ratelimiter"
)

func RateLimitMiddleware(ipRateLimiter *ratelimiter.IPRateLimiter, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Invalid Ip", http.StatusInternalServerError)
			return
		}
		limiter := ipRateLimiter.GetLimiter(ip)
		if limiter.Allow() {
			next(w, r)
		} else {
			http.Error(w, "Rate Limit Exceeded", http.StatusTooManyRequests)
		}
	}
}

func handleRequest(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "Request processed successfully at %v\n", time.Now())
}

func main() {
	ipRateLimiter := ratelimiter.NewIPRateLimiter()

	mux := http.NewServeMux()
	mux.HandleFunc("/", RateLimitMiddleware(ipRateLimiter, handleRequest))

	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", mux)
}
