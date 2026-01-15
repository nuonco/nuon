package main

import (
	"sync"
	"testing"
)

// BadUsage_UnlockInLoop should trigger the rule
func BadUsage_UnlockInLoop(t *testing.T) {
	var mu sync.Mutex
	// BUG: Unlock inside loop without defer; pairing issues
	for i := 0; i < 10; i++ {
		mu.Lock()
		mu.Unlock()
	}
}

// BadUsage_UnlockInRangeLoop should trigger the rule
func BadUsage_UnlockInRangeLoop(t *testing.T) {
	var mu sync.Mutex
	items := []int{1, 2, 3}
	// BUG: Unlock inside loop
	for _, item := range items {
		mu.Lock()
		_ = item
		mu.Unlock()
	}
}

// GoodUsage_DeferOutsideMutexLoop should NOT trigger the rule
func GoodUsage_DeferOutsideMutexLoop(t *testing.T) {
	var mu sync.Mutex
	// OK: defer ensures proper pairing
	mu.Lock()
	defer mu.Unlock()
	for i := 0; i < 10; i++ {
		_ = i
	}
}

// GoodUsage_LockDeferInsideFunctionCall should NOT trigger the rule
func GoodUsage_LockDeferInsideFunctionCall(t *testing.T) {
	// OK: Lock/Unlock in different scope
	for i := 0; i < 10; i++ {
		protectedWork()
	}
}

func protectedWork() {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()
}
