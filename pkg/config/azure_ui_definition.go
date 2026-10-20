package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// UploadedUIDefinition is an optional Azure createUiDefinition fetched from
// stack.toml and uploaded next to the stack templates. Path is what the vendor
// wrote. Contents is cleared once the body is in object storage.
type UploadedUIDefinition struct {
	Path         string                  `json:"path,omitempty"`
	Contents     string                  `json:"contents,omitempty"`
	ContentsHash string                  `json:"contents_hash,omitempty"`
	SourceURL    string                  `json:"source_url,omitempty"`
	Status       CustomNestedStackStatus `json:"status,omitempty"`
}

// ValidateAzureUIDefinition is the CLI check. The field is optional. When it is
// set, the stack must be azure-bicep. The document itself is checked later, when
// the config syncer uploads it.
func ValidateAzureUIDefinition(stackType, path string) error {
	if path == "" {
		return nil
	}
	if stackType != "azure-bicep" {
		msg := fmt.Sprintf("ui_definition is only supported when type is azure-bicep, got %q", stackType)
		return ErrConfig{Description: msg, Err: fmt.Errorf("%s", msg)}
	}
	return nil
}

// ValidateCreateUIDefinition checks a fetched createUiDefinition body. Empty
// contents are allowed: the file is optional, and callers that have not fetched
// it yet skip the check.
func ValidateCreateUIDefinition(contents string) error {
	if strings.TrimSpace(contents) == "" {
		return nil
	}
	if err := validateCreateUIDefinition(contents); err != nil {
		msg := fmt.Sprintf("ui_definition: %s", err)
		return ErrConfig{Description: msg, Err: fmt.Errorf("%s", msg)}
	}
	return nil
}

func validateCreateUIDefinition(contents string) error {
	var doc struct {
		Schema     string `json:"$schema"`
		Handler    string `json:"handler"`
		Parameters *struct {
			Basics json.RawMessage `json:"basics"`
			Steps  json.RawMessage `json:"steps"`
		} `json:"parameters"`
	}
	if err := json.Unmarshal([]byte(contents), &doc); err != nil {
		return fmt.Errorf("is not valid JSON: %w", err)
	}
	if doc.Schema == "" || doc.Handler == "" || doc.Parameters == nil {
		return fmt.Errorf("must be a createUiDefinition document with $schema, handler, and parameters")
	}
	if !isJSONArray(doc.Parameters.Basics) && !isJSONArray(doc.Parameters.Steps) {
		return fmt.Errorf("must include parameters.basics or parameters.steps")
	}
	return nil
}

func isJSONArray(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) > 0 && trimmed[0] == '['
}
