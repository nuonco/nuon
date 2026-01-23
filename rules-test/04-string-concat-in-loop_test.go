package main

import (
	"strings"
	"testing"
)

// BadUsage_StringConcatWithPlus should trigger the rule
func BadUsage_StringConcatWithPlus(t *testing.T) {
	// BUG: string concatenation in loop using +
	result := ""
	items := []string{"a", "b", "c"}
	for _, item := range items {
		result = result + item
	}
}

// BadUsage_StringConcatWithPlusEquals should trigger the rule
func BadUsage_StringConcatWithPlusEquals(t *testing.T) {
	// BUG: string concatenation in loop using +=
	output := ""
	for i := 0; i < 10; i++ {
		output += "x"
	}
}

// BadUsage_StringConcatInCLoop should trigger the rule
func BadUsage_StringConcatInCLoop(t *testing.T) {
	// BUG: string concatenation in C-style loop
	s := ""
	for i := 0; i < 100; i++ {
		s = s + "item"
	}
}

// BadUsage_StringConcatMultiLine should trigger the rule
func BadUsage_StringConcatMultiLine(t *testing.T) {
	// BUG: string concatenation with assignment in loop
	str := ""
	for _, name := range []string{"alice", "bob"} {
		str = str + "- " + name + "\n"
	}
}

// GoodUsage_StringBuilderInLoop should NOT trigger the rule
func GoodUsage_StringBuilderInLoop(t *testing.T) {
	// OK: using strings.Builder
	var builder strings.Builder
	items := []string{"a", "b", "c"}
	for _, item := range items {
		builder.WriteString(item)
	}
	_ = builder.String()
}

// GoodUsage_SliceJoinOutsideLoop should NOT trigger the rule
func GoodUsage_SliceJoinOutsideLoop(t *testing.T) {
	// OK: collecting in slice, joining after loop
	items := []string{"a", "b", "c"}
	var parts []string
	for _, item := range items {
		parts = append(parts, item)
	}
	_ = strings.Join(parts, ",")
}

// GoodUsage_NoStringConcat should NOT trigger the rule
func GoodUsage_NoStringConcat(t *testing.T) {
	// OK: no concatenation in loop
	output := ""
	for _, c := range "hello" {
		_ = output
		_ = c
	}
}

// GoodUsage_ConcatOutsideLoop should NOT trigger the rule
func GoodUsage_ConcatOutsideLoop(t *testing.T) {
	// OK: concatenation happens outside loop
	items := []string{"a", "b"}
	prefix := "start"
	for _, item := range items {
		_ = item
	}
	_ = prefix + "end"
}
