package models

import (
	"bufio"
	"net/url"
	"path"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/nuonco/nuon/pkg/config/schema"
)

func DetectSchemaType(text string) string {
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			comment := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			if comment != "" {
				return comment
			}
		}

		if !strings.HasPrefix(line, "#") {
			break
		}
	}
	return ""
}

func DetectSchemaTypeForDocument(text, uri string) string {
	if schemaType := DetectSchemaType(text); schemaType != "" {
		return schemaType
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		return ""
	}
	if strings.EqualFold(path.Base(parsed.Path), "branch.toml") {
		return "branch"
	}
	return ""
}

func LookupSchema(schemaType string) (*jsonschema.Schema, error) {
	return schema.LookupSchemaType(schemaType)
}

func GetValidSchemaTypes() []string {
	return schema.GetSchemaTypes()
}

func IsValidSchemaType(schemaType string) bool {
	return schema.IsValidSchemaType(schemaType)
}
