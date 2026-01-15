package main

import (
	"sync"
	"testing"
)

// BadUsage_DoneOutsideGoroutine should trigger the rule
func BadUsage_DoneOutsideGoroutine(t *testing.T) {
	var wg sync.WaitGroup
	// BUG: Done called in same scope as Add, not in goroutine
	wg.Add(1)
	wg.Done()
}

// BadUsage_DoneInParentScope should trigger the rule
func BadUsage_DoneInParentScope(t *testing.T) {
	var wg sync.WaitGroup
	// BUG: Done should be in goroutine
	wg.Add(1)
	go func() {
		// work
	}()
	wg.Done()
}

// GoodUsage_DoneInsideGoroutine should NOT trigger the rule
func GoodUsage_DoneInsideGoroutine(t *testing.T) {
	var wg sync.WaitGroup
	// OK: Done called inside goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		// work
	}()
	wg.Wait()
}

// GoodUsage_DoneWithDefer should NOT trigger the rule
func GoodUsage_DoneWithDefer(t *testing.T) {
	var wg sync.WaitGroup
	// OK: Done called inside goroutine with defer
	wg.Add(1)
	go func() {
		defer wg.Done()
	}()
	wg.Wait()
}
