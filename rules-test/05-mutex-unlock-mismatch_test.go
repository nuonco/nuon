package main

import (
	"sync"
	"testing"
)

// BadUsage_LockInBranch should trigger the rule
func BadUsage_LockInBranch(t *testing.T) {
	var mu sync.Mutex
	// BUG: Lock called in branch without defer
	if true {
		mu.Lock()
	}
}

// BadUsage_LockInLoop should trigger the rule
func BadUsage_LockInLoop(t *testing.T) {
	var mu sync.Mutex
	// BUG: Lock called in loop without defer
	for i := 0; i < 10; i++ {
		mu.Lock()
	}
}

// BadUsage_NoUnlockAtAll should trigger the rule
func BadUsage_NoUnlockAtAll(t *testing.T) {
	var mu sync.Mutex
	// BUG: Lock called but Unlock never called
	if true {
		mu.Lock()
		doSomething()
	}
}

// GoodUsage_DeferUnlock should NOT trigger the rule
func GoodUsage_DeferUnlock(t *testing.T) {
	var mu sync.Mutex
	// OK: defer ensures Unlock is called
	mu.Lock()
	defer mu.Unlock()
}

func doSomething() {}
