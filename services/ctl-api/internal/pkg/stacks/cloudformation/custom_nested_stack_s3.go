package cloudformation

import (
	"fmt"
	"strings"
)

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

func CustomNestedStackTemplateURL(baseURL, orgID, appID, contentsHash, templateURL string) string {
	key := CustomNestedStackS3Key(orgID, appID, contentsHash, templateURL)
	return fmt.Sprintf("%s/%s", strings.TrimSuffix(baseURL, "/"), key)
}
