package ratelimit

import (
	"sync"
	"time"
)

type Bucket struct {
	mu         sync.Mutex
	limit      int
	window     time.Duration
	tokens     int
	lastRefill time.Time
}

func NewBucket(limit int, window time.Duration) *Bucket {
	if limit <= 0 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}
	return &Bucket{limit: limit, window: window, tokens: limit, lastRefill: time.Now()}
}

func (b *Bucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	if now.Sub(b.lastRefill) >= b.window {
		b.tokens = b.limit
		b.lastRefill = now
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

type Registry struct {
	mu      sync.Mutex
	buckets map[string]*Bucket
}

func NewRegistry() *Registry {
	return &Registry{buckets: map[string]*Bucket{}}
}

func (r *Registry) Allow(jobID string, perMinute int) bool {
	r.mu.Lock()
	b, ok := r.buckets[jobID]
	if !ok {
		b = NewBucket(perMinute, time.Minute)
		r.buckets[jobID] = b
	}
	r.mu.Unlock()
	return b.Allow()
}
