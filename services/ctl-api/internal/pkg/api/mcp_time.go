package api

import "time"

const MCPTimeInstructions = "Timestamps in tool JSON are UTC RFC3339 ending in Z (Zulu). Convert each instant to the user's local timezone before naming a calendar day or clock time. Do not say today or yesterday from the UTC date."

func MCPTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func MCPTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return MCPTime(*t)
}
