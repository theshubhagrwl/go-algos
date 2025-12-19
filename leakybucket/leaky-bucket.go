package main

import (
	"fmt"
	"sync"
	"time"
)

type LeakyBucket struct {
	size        int
	water       int
	outflowRate int
	lastLeak    time.Time
	mu          sync.Mutex
}

func NewLeakyBucket(size int, outflowRate int) *LeakyBucket {
	return &LeakyBucket{
		size:        size,
		water:       0,
		outflowRate: outflowRate,
		lastLeak:    time.Now(),
	}
}

func (lb *LeakyBucket) leak() {
	now := time.Now()
	elapsed := now.Sub(lb.lastLeak).Seconds()
	leaked := elapsed * float64(lb.outflowRate)

	lb.water -= int(leaked)
	if lb.water < 0 {
		lb.water = 0
	}
	lb.lastLeak = now
}

func (lb *LeakyBucket) Allow(n int) bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.leak()
	if lb.water+n <= lb.size {
		lb.water += n
		return true
	}
	return false
}

func (lb *LeakyBucket) WaterLevel() int {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.leak()
	return lb.water
}

func main() {
	fmt.Println("Leaky Bucket Algorithm")

	bucket := NewLeakyBucket(5, 5)

	for i := 1; i <= 15; i++ {
		if bucket.Allow(1) {
			fmt.Printf("Request %2d: ✓ ALLOWED  - Water level: %2d\n", i, bucket.WaterLevel())
		} else {
			fmt.Printf("Request %2d: ✗ DENIED   - Water level: %2d (FULL!)\n", i, bucket.WaterLevel())
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("\nWaiting 3 seconds for bucket to leak...")
	time.Sleep(3 * time.Second)

	fmt.Println("\nTest 2: After bucket has leaked")
	for i := 16; i <= 20; i++ {
		if bucket.Allow(1) {
			fmt.Printf("Request %2d: ✓ ALLOWED  - Water level: %2d\n", i, bucket.WaterLevel())
		} else {
			fmt.Printf("Request %2d: ✗ DENIED   - Water level: %2d\n", i, bucket.WaterLevel())
		}
	}

}
