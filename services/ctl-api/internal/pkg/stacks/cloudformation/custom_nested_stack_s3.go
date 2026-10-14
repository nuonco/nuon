package cloudformation

import (
	"fmt"
	"strings"
)

// CustomNestedStackS3Key returns the S3 object key for a custom nested stack template.
func CustomNestedStackS3Key(orgID, appID, contentsHash, templateURL string) string {
	ext := ".yaml"
	lower := strings.ToLower(templateURL)
	if strings.HasSuffix(lower, ".json") {
		ext = ".json"
	} else if strings.HasSuffix(lower, ".yml") {
		ext = ".yml"
	}
	return fmt.Sprintf("stacks/%s/%s/%s%s", orgID, appID, contentsHash, ext)
}

// CustomNestedStackTemplateURL returns the full S3 HTTPS URL for the template.
func CustomNestedStackTemplateURL(baseURL, orgID, appID, contentsHash, templateURL string) string {
	key := CustomNestedStackS3Key(orgID, appID, contentsHash, templateURL)
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(baseURL, "/"), key)
}

// UIDefinitionS3Key returns the object key for an uploaded Azure createUiDefinition.
func UIDefinitionS3Key(orgID, appID, contentsHash string) string {
	return fmt.Sprintf("stacks/%s/%s/%s.ui.json", orgID, appID, contentsHash)
}

// UIDefinitionSourceURL returns the HTTPS URL for an uploaded createUiDefinition.
func UIDefinitionSourceURL(baseURL, orgID, appID, contentsHash string) string {
	if baseURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(baseURL, "/"), UIDefinitionS3Key(orgID, appID, contentsHash))
}
