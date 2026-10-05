package appbundle

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const InputPlaceholderPrefix = "__NUON_INPUT_"

func InputPlaceholder(name string) string {
	return InputPlaceholderPrefix + name + "__"
}

const componentOutputPlaceholderPrefix = "__NUON_CUSTOMER_MANAGED_COMPONENT_"

func ComponentOutputPlaceholder(componentName, outputPath string) string {
	sum := sha256.Sum256([]byte(componentName + "\x00" + outputPath))
	sanitized := strings.NewReplacer(".", "_", "-", "_").Replace(componentName + "_" + outputPath)
	return componentOutputPlaceholderPrefix + sanitized + "_" + hex.EncodeToString(sum[:])[:8] + "__"
}

// ParseInputPlaceholder returns the input name encoded by an exact
// InputPlaceholder token, and whether value is one.
func ParseInputPlaceholder(value string) (string, bool) {
	if len(value) < len(InputPlaceholderPrefix)+3 || !strings.HasPrefix(value, InputPlaceholderPrefix) || !strings.HasSuffix(value, "__") {
		return "", false
	}
	return strings.TrimSuffix(strings.TrimPrefix(value, InputPlaceholderPrefix), "__"), true
}
