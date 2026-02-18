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
