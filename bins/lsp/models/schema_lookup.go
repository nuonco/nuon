package models

import (
	"bufio"
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

func LookupSchema(schemaType string) (*jsonschema.Schema, error) {
	return schema.LookupSchemaType(schemaType)
}
