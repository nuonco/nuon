package main

import (
	"fmt"
	"testing"
)

// BadUsage_DeferInForLoop should trigger the rule
func BadUsage_DeferInForLoop(t *testing.T) {
	for i := 0; i < 3; i++ {
		defer fmt.Println(i) // BUG: defers until end of function, not each iteration
	}
}

// BadUsage_DeferInRangeLoop should trigger the rule
func BadUsage_DeferInRangeLoop(t *testing.T) {
	items := []string{"a", "b", "c"}

	for _, item := range items {
		defer fmt.Println(item) // BUG: all defers execute at end of function
	}
}

// BadUsage_DeferInMapRange should trigger the rule
func BadUsage_DeferInMapRange(t *testing.T) {
	data := map[string]int{"x": 1, "y": 2}

	for key, val := range data {
		defer fmt.Printf("%s: %d\n", key, val) // BUG: defers stack up
	}
}

// GoodUsage_NoDefer should NOT trigger the rule
func GoodUsage_NoDefer(t *testing.T) {
	for i := 0; i < 3; i++ {
		fmt.Println(i) // OK: no defer
	}
}

// GoodUsage_DeferOutsideLoop should NOT trigger the rule
func GoodUsage_DeferOutsideLoop(t *testing.T) {
	defer fmt.Println("cleanup") // OK: defer is outside loop

	for i := 0; i < 3; i++ {
		fmt.Println(i)
	}
}

// GoodUsage_RangeWithoutDefer should NOT trigger the rule
func GoodUsage_RangeWithoutDefer(t *testing.T) {
	items := []string{"a", "b", "c"}

	for _, item := range items {
		fmt.Println(item) // OK: no defer in loop
	}
}
