package ratelimiter

import (
	"math"
	"sync"
	"time"
)

type TokenBucketRateLimiter struct {
	capacity       float64 `json:"capacity"`       // 桶的容量
	tokens         float64 `json:"tokens"`         // 当前桶中的令牌数
	genTokenRate   float64 `json:"genTokenRate"`   // 每秒生成的令牌速率
	lastGenTokenMs int64   `json:"lastGenTokenMs"` // 上次生成令牌的毫秒数
	lock           sync.Mutex
}

func NewTokenBucketRateLimiter(genTokenRate, capacity float64) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		genTokenRate:   genTokenRate,
		capacity:       capacity,
		tokens:         0,
		lastGenTokenMs: time.Now().UnixNano() / 1e6,
	}
}

func (l *TokenBucketRateLimiter) Allow() bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	now := time.Now().UnixNano() / 1e6
	// 计算两个时间内生成的令牌数
	tokens := (float64(now - l.lastGenTokenMs)) * l.genTokenRate / 1000
	// 计算当前桶内令牌数
	l.tokens = math.Min(l.capacity, l.tokens+tokens)
	l.lastGenTokenMs = now
	if l.tokens > 0 {
		// 获取到令牌，允许执行
		l.tokens--
		return true
	} else {
		// 没有令牌，不允许执行
		return false
	}
}
