package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type SlidingWindowLog struct {
	limit  int
	windowSize  time.Duration 
	requestLog  []time.Time
	mu sync.Mutex 
}

func NewSlidingWindow(limit int, windowSize time.Duration) *SlidingWindowLog {
	return &SlidingWindowLog {
		limit : limit,
		windowSize : windowSize,
		requestLog : make([]time.Time,0),
	}
}

func (swl *SlidingWindowLog) removeOldRequests(now time.Time){
	cutoff := now.Add(-swl.windowSize)

	validIndex := 0
	for validIndex < len(swl.requestLog) && swl.requestLog[validIndex].Before(cutoff) {
		validIndex++
	}

	if validIndex > 0 {
		swl.requestLog = swl.requestLog[validIndex:]
	}
}

func (swl* SlidingWindowLog) Allow() bool {
	swl.mu.Lock()
	defer swl.mu.Unlock()

	now := time.Now()

	swl.removeOldRequests(now);

	if len(swl.requestLog) < swl.limit {
		swl.requestLog = append(swl.requestLog,now)
		return true
	}
	return false
}

func (swl *SlidingWindowLog) RequestInWindow() int {
	swl.mu.Lock()
	defer swl.mu.Unlock()

	swl.removeOldRequests(time.Now())
	return len(swl.requestLog)
}

func (swl *SlidingWindowLog) RemainingInWindow() int {
	swl.mu.Lock()
	swl.mu.Unlock()

	swl.removeOldRequests(time.Now())
	return swl.limit - len(swl.requestLog)
}

func main(){
	window := NewSlidingWindow(5,2 * time.Second)

	for i := 1; i <= 10; i++ {
		if window.Allow() {
			fmt.Printf("Request %2d: ✓ ALLOWED  - Count: %d/%d, Remaining: %d\n",
		i, window.RequestInWindow(), 5, window.RemainingInWindow())
		} else {
			fmt.Printf("Request %2d: ✗ DENIED   - Count: %d/%d (window full!)\n",
				i, window.RequestInWindow(), 5)
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("TEST 2: Wait for window to slide (3 seconds)...")
	fmt.Println(strings.Repeat("─", 70))
	time.Sleep(3 * time.Second)
	fmt.Printf("\nRequests in window after wait: %d (old requests expired!)\n\n", window.RemainingInWindow())



	// Create a fresh limiter for this demo
	boundaryLimiter := NewSlidingWindow(5, 1 * time.Second)
	
	fmt.Println("Scenario: 1 second window, 5 requests limit")
	fmt.Println()
	
	// Make 5 requests at end of first second
	fmt.Println("At t=0.9s: Make 5 requests")
	time.Sleep(900 * time.Millisecond)
	for i := 1; i <= 5; i++ {
		boundaryLimiter.Allow()
	}
	fmt.Printf("  Requests in window: %d/%d ✓\n", boundaryLimiter.RequestInWindow(), 5)
	
	// Try to make 5 more at start of next second
	fmt.Println("\nAt t=1.1s: Try to make 5 more requests")
	time.Sleep(200 * time.Millisecond)
	
	allowed := 0
	denied := 0
	for i := 1; i <= 5; i++ {
		if boundaryLimiter.Allow() {
			allowed++
		} else {
			denied++
		}
	}
	
	fmt.Printf("  Allowed: %d, Denied: %d\n", allowed, denied)
	fmt.Printf("  Requests in window: %d/%d\n", boundaryLimiter.RequestInWindow(), 5)
}
