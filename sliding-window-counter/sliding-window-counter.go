package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type SlidingWindowCounter struct {
	limit int 
	windowSize time.Duration
	previousCount int 
	currentCount int 
	currentWindowStart time.Time 
	mu sync.Mutex
}

func NewSlidingWindowCounter (limit int, windowSize time.Duration) *SlidingWindowCounter {
	return &SlidingWindowCounter{
		limit:limit, 
		windowSize: windowSize,
		previousCount:  0, 
		currentCount: 0, 
		currentWindowStart: time.Now(),
	}
}

func (swc *SlidingWindowCounter) slideWindow(now time.Time) {
	elapsed := now.Sub(swc.currentWindowStart)

	if elapsed >= swc.windowSize{
		swc.previousCount = swc.currentCount
		swc.currentCount = 0
		swc.currentWindowStart = now
	}
}

func (swc *SlidingWindowCounter) getWeightedCount(now time.Time) float64 {
	elapsed := now.Sub(swc.currentWindowStart).Seconds()
	windowSizeSeconds := swc.windowSize.Seconds()

	currentWindowProgress := elapsed/windowSizeSeconds
	if currentWindowProgress > 1.0 {
		currentWindowProgress = 1.0
	}

	previousWindowWeight := 1.0 - currentWindowProgress

	weightedCount := float64(swc.previousCount) * previousWindowWeight + float64(swc.currentCount)
	return weightedCount
}

func (swc *SlidingWindowCounter) Allow() bool {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	now := time.Now()

	swc.slideWindow(now)

	weightedCount := swc.getWeightedCount(now)
	if weightedCount < float64(swc.limit) {
		swc.currentCount++
		return true 
	}
	return false
}

func (swc *SlidingWindowCounter) RequestsInWindow() float64 {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	swc.slideWindow(time.Now())
	return swc.getWeightedCount((time.Now()))
}

func (swc *SlidingWindowCounter) RemainingInWindow() float64 {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	swc.slideWindow(time.Now())
	weightedCount := swc.getWeightedCount(time.Now())
	remaining := float64(swc.limit) - weightedCount
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

func (swc *SlidingWindowCounter) GetCounts() (int, int) {
	swc.mu.Lock()
	defer swc.mu.Unlock()

	swc.slideWindow(time.Now())
	return swc.previousCount, swc.currentCount
}

func main() {
	limiter := NewSlidingWindowCounter(5,2*time.Second)

	fmt.Println("TEST 1: Rapid requests (10 requests quickly)")

	for i := 1; i <= 10; i++ {
		prev, curr := limiter.GetCounts()
		if limiter.Allow() {
			fmt.Printf("Request %2d: ✓ ALLOWED  - Weighted: %.1f/5, [Prev:%d, Curr:%d]\n",
				i, limiter.RequestsInWindow(), prev, curr)
		} else {
			fmt.Printf("Request %2d: ✗ DENIED   - Weighted: %.1f/5 (limit!)\n",
				i, limiter.RequestsInWindow())
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("TEST 2: Wait for new window (3 seconds)...")
	fmt.Println(strings.Repeat("─", 70))
	time.Sleep(3 * time.Second)
	prev, curr := limiter.GetCounts()
	fmt.Printf("\nAfter wait - Prev: %d, Curr: %d, Weighted: %.1f\n\n", 
		prev, curr, limiter.RequestsInWindow())
	

		fmt.Println("TEST 3: After window slide")
	fmt.Println(strings.Repeat("─", 70))
	fmt.Println()
	
	for i := 11; i <= 15; i++ {
		prev, curr := limiter.GetCounts()
		if limiter.Allow() {
			fmt.Printf("Request %2d: ✓ ALLOWED  - Weighted: %.1f/5, [Prev:%d, Curr:%d]\n",
				i, limiter.RequestsInWindow(), prev, curr)
		} else {
			fmt.Printf("Request %2d: ✗ DENIED   - Weighted: %.1f/5\n",
				i, limiter.RequestsInWindow())
		}
	}

}