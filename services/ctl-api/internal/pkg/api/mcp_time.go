package api

import "time"

// MCPTimeInstructions is included in MCP server instructions so clients
// treat timestamps as UTC and localize before naming a day or clock time.
const MCPTimeInstructions = "Timestamps in tool JSON are UTC RFC3339 ending in Z (Zulu). Convert each instant to the user's local timezone before naming a calendar day or clock time. Do not say today or yesterday from the UTC date."

// MCPTime formats a timestamp for MCP payloads as UTC RFC3339 (Zulu, …Z).
// Agents localize this instant before naming a calendar day or clock time.
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
