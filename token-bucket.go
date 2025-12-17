package main

import (
	"fmt"
	"sync"
	"time"
)

type TokenBucket struct {
	capacity   int
	tokens     int
	refillRate int
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucket(capacity int, refillRate int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()

	tokensToAdd := elapsed * float64(tb.refillRate)
	tb.tokens += int(tokensToAdd)

	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	tb.lastRefill = now
}

func (tb *TokenBucket) Allow(n int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}
	return false
}

func (tb *TokenBucket) AvailableTokens() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	return tb.tokens
}

func main() {
	fmt.Println("Token Bucket Algorithm")

	bucket := NewTokenBucket(10, 2)

	for i := 1; i <= 15; i++ {
		if bucket.Allow(1) {
			fmt.Printf("Request %2d Allowed - Tokens left %2d \n", i, bucket.AvailableTokens())
		} else {
			fmt.Printf("Request %2d Denied - Tokens left %2d \n", i, bucket.AvailableTokens())
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("wait 3 sec to refill")
	time.Sleep(3 * time.Second)

	for i := 16; i <= 20; i++ {
		if bucket.Allow(1) {
			fmt.Printf("Request %2d Allowed - Token left %2d \n", i, bucket.AvailableTokens())
		} else {
			fmt.Printf("Request %2d Denied - Token left %2d \n", i, bucket.AvailableTokens())
		}
	}

}
