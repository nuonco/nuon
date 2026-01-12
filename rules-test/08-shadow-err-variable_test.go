package main

import (
	"testing"
)

// BadUsage_ShadowErrInIf should trigger the rule
func BadUsage_ShadowErrInIf(t *testing.T) {
	// BUG: err shadowed in inner if scope
	err := doSomethingWithError()
	if true {
		err := anotherError()
		_ = err
	}
	_ = err
}

// BadUsage_ShadowErrInFor should trigger the rule
func BadUsage_ShadowErrInFor(t *testing.T) {
	// BUG: err shadowed in for loop
	err := doSomethingWithError()
	for i := 0; i < 1; i++ {
		err := anotherError()
		_ = err
	}
	_ = err
}

// BadUsage_ShadowErrInSwitch should trigger the rule
func BadUsage_ShadowErrInSwitch(t *testing.T) {
	// BUG: err shadowed in switch case
	err := doSomethingWithError()
	switch {
	case true:
		err := anotherError()
		_ = err
	}
	_ = err
}

// GoodUsage_NoShadowErr should NOT trigger the rule
func GoodUsage_NoShadowErr(t *testing.T) {
	// OK: single err declaration at function level
	err := doSomethingWithError()
	_ = err
}

// GoodUsage_ErrDetailsNotShadowed should NOT trigger the rule
func GoodUsage_ErrDetailsNotShadowed(t *testing.T) {
	// OK: err_details at function level, not shadowing
	err_details := doSomethingWithError()
	_ = err_details
}

// GoodUsage_ReassignNotShadow should NOT trigger the rule
func GoodUsage_ReassignNotShadow(t *testing.T) {
	// OK: using = instead of := is reassignment, not shadowing
	err := doSomethingWithError()
	if true {
		err = anotherError()
		_ = err
	}
	_ = err
}

func anotherError() error {
	return nil
}

func doSomethingWithError() error {
	return nil
}
