package ratelimiter

import (
	"math"
	"sync"
	"time"
)

type LeakyBucketRateLimiter struct {
	capacity   float64 `json:"capacity"`   // 桶的容量
	water      float64 `json:"water"`      // 当前桶中水量
	flowRate   float64 `json:"flowRate"`   // 每秒漏桶流速
	lastLeakMs int64   `json:"lastLeakMs"` // 上次漏水毫秒数
	lock       sync.Mutex
}

func NewLeakyBucketRateLimiter(flowRate, capacity float64) *LeakyBucketRateLimiter {
	return &LeakyBucketRateLimiter{
		capacity:   capacity,
		flowRate:   flowRate,
		water:      capacity + 1, // 处于一个新的开始
		lastLeakMs: time.Now().UnixNano() / 1e6,
	}
}

func (l *LeakyBucketRateLimiter) Allow() bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	// 获取当前时间
	now := time.Now().UnixNano() / 1e6
	// 计算这段时间流出的水量：
	outflowWater := (float64(now - l.lastLeakMs)) * l.flowRate / 1000
	// 计算水量： 桶的当前水量 - 流出的水量
	l.water = math.Max(0, l.water-outflowWater)
	l.lastLeakMs = now
	if l.water < l.capacity {
		// 当前水量 小于 桶容量，允许通过
		l.water++
		return true
	} else {
		// 当前水量 不小于 桶容量，不允许通过
		return false
	}
}
