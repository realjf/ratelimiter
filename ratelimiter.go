package ratelimiter

type RateLimiter interface {
	Allow() bool
}
