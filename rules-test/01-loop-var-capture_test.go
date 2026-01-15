package main

import (
	"fmt"
	"sync"
	"testing"
)

// BadUsage_CapturesLoopVar should trigger the rule
func BadUsage_CapturesLoopVar(t *testing.T) {
	items := []int{1, 2, 3}

	wg := sync.WaitGroup{}
	for i := range items {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println(i) // BUG: captures loop variable
		}()

	}
}

// BadUsage_CapturesLoopValue should trigger the rule
func BadUsage_CapturesLoopValue(t *testing.T) {
	data := []string{"a", "b", "c"}

	for _, val := range data {
		go func() {
			fmt.Println(val) // BUG: captures loop variable
		}()
	}
}

// GoodUsage_PassAsArg should NOT trigger the rule
func GoodUsage_PassAsArg(t *testing.T) {
	items := []int{1, 2, 3}

	for i := range items {
		go func(index int) {
			fmt.Println(index)
		}(i)
	}
}

// GoodUsage_ShadowVariable should NOT trigger the rule
func GoodUsage_ShadowVariable(t *testing.T) {
	data := []string{"a", "b", "c"}

	for _, val := range data {
		val := val // Shadow the loop variable
		go func() {
			fmt.Println(val)
		}()
	}
}

// GoodUsage_NoGoroutine should NOT trigger the rule
func GoodUsage_NoGoroutine(t *testing.T) {
	items := []int{1, 2, 3}

	for i := range items {
		fmt.Println(i) // No goroutine, no issue
	}
}
