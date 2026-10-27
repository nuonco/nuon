package cloudformation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/pkg/config"
)

// The template bucket is publicly readable, so object keys must not reveal org or
// app ids. Keys are derived from the ids and the contents hash: stable for a given
// upload, but not reversible to the ids that produced them.
func stackObjectID(orgID, appID, contentsHash string) string {
	sum := sha256.Sum256([]byte(orgID + "/" + appID + "/" + contentsHash))
	return hex.EncodeToString(sum[:])
}

func templateExtension(templateURL string) string {
	lower := strings.ToLower(templateURL)
	switch {
	case strings.HasSuffix(lower, ".json"):
		return ".json"
	case strings.HasSuffix(lower, ".yml"):
		return ".yml"
	default:
		return ".yaml"
	}
}

func joinBaseURL(baseURL, key string) string {
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(baseURL, "/"), key)
}

// CustomNestedStackS3Key returns the S3 object key for a custom nested stack template.
func CustomNestedStackS3Key(orgID, appID, contentsHash, templateURL string) string {
	return fmt.Sprintf("stacks/%s%s", stackObjectID(orgID, appID, contentsHash), templateExtension(templateURL))
}

// CustomNestedStackTemplateURL returns the full S3 HTTPS URL for the template.
func CustomNestedStackTemplateURL(baseURL, orgID, appID, contentsHash, templateURL string) string {
	return joinBaseURL(baseURL, CustomNestedStackS3Key(orgID, appID, contentsHash, templateURL))
}

// legacyCustomNestedStackTemplateURL is where templates uploaded before keys were
// made opaque still live. Configs from that era may not have TemplateSourceURL set.
func legacyCustomNestedStackTemplateURL(baseURL, orgID, appID, contentsHash, templateURL string) string {
	return joinBaseURL(baseURL, fmt.Sprintf("stacks/%s/%s/%s%s", orgID, appID, contentsHash, templateExtension(templateURL)))
}

// ResolveCustomNestedStackTemplateURL returns the URL a custom nested stack
// template is fetched from: the recorded upload URL when present, the legacy
// layout for uploads that predate it, and the vendor's URL when never uploaded.
func ResolveCustomNestedStackTemplateURL(baseURL, orgID, appID string, stack config.CustomNestedStack) string {
	if stack.TemplateSourceURL != "" {
		return stack.TemplateSourceURL
	}
	if stack.ContentsHash != "" && baseURL != "" {
		return legacyCustomNestedStackTemplateURL(baseURL, orgID, appID, stack.ContentsHash, stack.TemplateURL)
	}
	return stack.TemplateURL
}

// UIDefinitionS3Key returns the object key for an uploaded Azure createUiDefinition.
func UIDefinitionS3Key(orgID, appID, contentsHash string) string {
	return fmt.Sprintf("stacks/%s.ui.json", stackObjectID(orgID, appID, contentsHash))
}

// UIDefinitionSourceURL returns the HTTPS URL for an uploaded createUiDefinition.
func UIDefinitionSourceURL(baseURL, orgID, appID, contentsHash string) string {
	if baseURL == "" {
		return ""
	}
	return joinBaseURL(baseURL, UIDefinitionS3Key(orgID, appID, contentsHash))
}

func legacyUIDefinitionSourceURL(baseURL, orgID, appID, contentsHash string) string {
	if baseURL == "" {
		return ""
	}
	return joinBaseURL(baseURL, fmt.Sprintf("stacks/%s/%s/%s.ui.json", orgID, appID, contentsHash))
}
