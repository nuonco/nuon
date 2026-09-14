package customermanaged

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
