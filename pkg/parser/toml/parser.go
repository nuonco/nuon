package toml

// ParseToml parses TOML text using loose parsing for LSP
// Always uses loose parsing to preserve position information
// Never returns an error - the LSP must remain operational for all inputs
func ParseToml(text string) *TomlDocument {
	// Use loose parsing to preserve line/column information
	// Strict parsing loses position data which is critical for LSP
	return ParseLoose(text)
}

// ParseTomlWithCursor parses TOML and provides cursor context
// Uses loose parsing with cursor-aware position tracking
func ParseTomlWithCursor(text string, cursorPos Position) *TomlDocument {
	// Always use loose parsing for LSP to preserve positions
	return ParseLooseWithCursor(text, cursorPos)
}

// ValidateToml validates TOML syntax using strict parser
// Returns nil if valid, error if invalid
// Use this for validation/diagnostics, not for position-based features
func ValidateToml(text string) error {
	_, err := ParseStrict(text)
	return err
}
