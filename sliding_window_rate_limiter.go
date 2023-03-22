package ratelimiter

import (
	"sync"
	"time"
)

type slot struct {
	timestamp time.Time `json:"timestamp"`
	counter   int       `json:"counter"`
}

type SlidingWindowRateLimiter struct {
	lock         sync.Mutex
	numSlots     int           `json:"numSlots"`     // 子窗口数量
	threshold    int           `json:"threshold"`    // 阈值
	slotInterval time.Duration `json:"slotInterval"` // 子窗口时间长度
	winInterval  time.Duration `json:"winInterval"`  // 大窗口时间长度
	slots        []*slot       `json:"slots"`        // 子窗口切片
}

func NewSlidingWindowRateLimiter(slotInterval, winInterval time.Duration, threshold int) *SlidingWindowRateLimiter {
	numSlots := int(winInterval / slotInterval)
	return &SlidingWindowRateLimiter{
		numSlots:     numSlots,
		threshold:    threshold,
		slotInterval: slotInterval,
		winInterval:  winInterval,
	}
}

func (l *SlidingWindowRateLimiter) Allow() bool {
	l.lock.Lock()
	defer l.lock.Unlock()

	now := time.Now()
	// 已经过期的slot移出时间窗
	invalidOffset := -1
	for i, s := range l.slots {
		if s.timestamp.Add(l.winInterval).After(now) {
			break
		}
		invalidOffset = i
	}
	if invalidOffset > -1 {
		l.slots = l.slots[invalidOffset+1:]
	}

	// 判断请求是否达到阈值
	var allowed bool
	if l.count() < l.threshold {
		allowed = true
	}

	// 记录这次的请求
	lastSlot := &slot{}
	if len(l.slots) > 0 {
		lastSlot = l.slots[len(l.slots)-1]
		if lastSlot.timestamp.Add(l.slotInterval).Before(now) {
			// 如果当前时间已经超过最后时间插槽的跨度，那么新建一个时间插槽
			lastSlot = &slot{timestamp: now, counter: 1}
			l.slots = append(l.slots, lastSlot)
		} else {
			lastSlot.counter++
		}
	} else {
		lastSlot = &slot{timestamp: now, counter: 1}
		l.slots = append(l.slots, lastSlot)
	}

	return allowed
}

func (l *SlidingWindowRateLimiter) count() int {
	count := 0
	for _, s := range l.slots {
		count += s.counter
	}
	return count
}
