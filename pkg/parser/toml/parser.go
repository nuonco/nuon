package toml

func ParseToml(text string) *TomlDocument {
	return ParseLoose(text)
}

func ParseTomlWithCursor(text string, cursorPos Position) *TomlDocument {
	return ParseLooseWithCursor(text, cursorPos)
}

func ValidateToml(text string) error {
	_, err := ParseStrict(text)
	return err
}
