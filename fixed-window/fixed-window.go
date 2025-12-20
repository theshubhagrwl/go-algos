package main

import (
	"fmt"
	"sync"
	"time"
)

type FixedWindow struct {
	limit       int
	windowSize  time.Duration
	counter     int
	windowStart time.Time
	mu          sync.Mutex
}

func NewFixedWindow(limit int, windowSize time.Duration) *FixedWindow {
	return &FixedWindow{
		limit:       limit,
		counter:     0,
		windowSize:  windowSize,
		windowStart: time.Now(),
	}
}

func (fw *FixedWindow) reset() {
	now := time.Now()
	elapsed := now.Sub(fw.windowStart)
	if elapsed >= fw.windowSize {
		fw.counter = 0
		fw.windowStart = now
	}
}

func (fw *FixedWindow) Allow() bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	fw.reset()

	if fw.counter < fw.limit {
		fw.counter++
		return true
	}
	return false
}

func (fw *FixedWindow) RequestInWindow() int {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	return fw.counter
}

func (fw *FixedWindow) RemainingWindow() int {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	return fw.limit - fw.counter
}

func main() {

	window := NewFixedWindow(5, 2*time.Second)

	for i := 1; i <= 10; i++ {
		if window.Allow() {
			fmt.Printf("Request %2d:  ALLOWED  - Count: %d/%d, Remaining: %d\n",
				i, window.RequestInWindow(), 5, window.RemainingWindow())
		} else {
			fmt.Printf("Request %2d:  DENIED   - Count: %d/%d (window limit reached!)\n",
				i, window.RequestInWindow(), 5)
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("Waiting for 3 seconds")
	time.Sleep(3 * time.Second)

	for i := 11; i <= 15; i++ {
		if window.Allow() {
			fmt.Printf("Request %2d:  ALLOWED  - Count: %d/%d, Remaining: %d\n",
				i, window.RequestInWindow(), 5, window.RemainingWindow())
		} else {
			fmt.Printf("Request %2d:  DENIED   - Count: %d/%d\n",
				i, window.RequestInWindow(), 5)
		}
	}

	// Window Boundary problem example
	boundaryCounter := NewFixedWindow(5, 1*time.Second)

	fmt.Println("At t=0.9s (end of window 1): Make 5 requests")
	time.Sleep(900 * time.Millisecond) // near end of window
	for i := 1; i <= 5; i++ {
		if boundaryCounter.Allow() {
			fmt.Printf("Request %2d:  ALLOWED  - Count: %d/%d, Remaining: %d\n",
				i, boundaryCounter.RequestInWindow(), 5, boundaryCounter.RemainingWindow())
		} else {
			fmt.Printf("Request %2d:  DENIED   - Count: %d/%d\n",
				i, boundaryCounter.RequestInWindow(), 5)
		}
	}
	// fmt.Printf("  Window 1 count: %d/%d ALLOWED\n", boundaryCounter.RequestInWindow(), 5)

	// Wait just a bit to cross into new window
	fmt.Println("\nAt t=1.1s (start of window 2): Make 5 more requests")
	time.Sleep(200 * time.Millisecond) // new window
	for i := 1; i <= 5; i++ {
		if boundaryCounter.Allow() {
			fmt.Printf("Request %2d:  ALLOWED  - Count: %d/%d, Remaining: %d\n",
				i, boundaryCounter.RequestInWindow(), 5, boundaryCounter.RemainingWindow())
		} else {
			fmt.Printf("Request %2d:  DENIED   - Count: %d/%d\n",
				i, boundaryCounter.RequestInWindow(), 5)
		}
	}
	// fmt.Printf("  Window 2 count: %d/%d ALLOWED\n", boundaryCounter.RequestInWindow(), 5)

}
