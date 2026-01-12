package main

import "testing"

// BadUsage_AssignmentToNilMap
// Assigning to an uninitialized map will cause runtime panic
func BadUsage_AssignmentToNilMap(t *testing.T) {
	_ = t
	var m map[string]int
	// BUG: this will panic - map not initialized
	m["key"] = 42
}

// BadUsage_NilMapAssignmentInFunction
func BadUsage_NilMapAssignmentInFunction(t *testing.T) {
	_ = t
	var data map[string]string
	// BUG: panic on nil map assignment
	data["field"] = "value"
}

// GoodUsage_MapInitializedWithMake
func GoodUsage_MapInitializedWithMake(t *testing.T) {
	_ = t
	m := make(map[string]int)
	// OK: map is initialized
	m["key"] = 42
}

// GoodUsage_MapInitializedWithLiteral
func GoodUsage_MapInitializedWithLiteral(t *testing.T) {
	_ = t
	m := map[string]int{"key": 42}
	// OK: map is initialized with literal
	m["key2"] = 100
}
