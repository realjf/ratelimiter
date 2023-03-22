package ratelimiter

import (
	"sync"
	"time"
)

type FixedWindowRateLimiter struct {
	threshold int           `json:"threshold"` // 阈值
	stime     time.Time     `json:"stime"`     // 开始时间
	interval  time.Duration `json:"interval"`  // 时间窗口
	counter   int           `json:"counter"`   // 当前计数
	lock      sync.Mutex
}

func NewFixedWindowRateLimiter(threshold int, interval time.Duration) *FixedWindowRateLimiter {

	return &FixedWindowRateLimiter{
		threshold: threshold,
		stime:     time.Now(),
		interval:  interval,
		counter:   threshold - 1, // 让其处于下一个时间窗口开始的时间临界点
	}
}

func (l *FixedWindowRateLimiter) Allow() bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	// 判断收到请求数是否达到阈值
	if l.counter == l.threshold-1 {
		now := time.Now()
		// 达到阈值后，判断是否是请求窗口内
		if now.Sub(l.stime) >= l.interval {
			// 重新计数
			l.Reset()
			return true
		}
		// 丢弃多余的请求
		return false
	} else {
		l.counter++
		return true
	}
}

func (l *FixedWindowRateLimiter) Reset() {
	l.counter = 0
	l.stime = time.Now()
}
