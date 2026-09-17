package models

import (
	"bufio"
	"net/url"
	"path"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/nuonco/nuon/pkg/config/schema"
)

const schemaDirectivePrefix = "#:schema"

// SchemaDeclaration is the schema type declared in a document's leading
// comment block, with the position of the type token for diagnostics.
type SchemaDeclaration struct {
	Type  string
	Line  int
	Start int
	End   int
}

func DetectSchemaType(text string) string {
	decl := DetectSchema(text)
	if decl == nil {
		return ""
	}
	return decl.Type
}

// DetectSchema scans the leading comment block for a `#:schema <url>`
// directive and resolves the type from its last path segment (or ?type=
// query). Without a directive, a bare `# <type>` comment is honored only when
// it names a known schema type, so prose comments are never misread.
func DetectSchema(text string) *SchemaDeclaration {
	var bare *SchemaDeclaration

	scanner := bufio.NewScanner(strings.NewReader(text))
	for lineNo := 0; scanner.Scan(); lineNo++ {
		raw := scanner.Text()
		line := strings.TrimSpace(raw)

		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "#") {
			break
		}

		if strings.HasPrefix(line, schemaDirectivePrefix) {
			ref := strings.TrimSpace(strings.TrimPrefix(line, schemaDirectivePrefix))
			if ref == "" {
				continue
			}
			typ := schemaTypeFromRef(ref)
			start := strings.Index(raw, ref)
			return &SchemaDeclaration{Type: typ, Line: lineNo, Start: start, End: start + len(ref)}
		}

		if bare != nil {
			continue
		}
		comment := strings.TrimSpace(strings.TrimPrefix(line, "#"))
		if comment != "" && schema.IsValidSchemaType(comment) {
			start := strings.Index(raw, comment)
			bare = &SchemaDeclaration{Type: comment, Line: lineNo, Start: start, End: start + len(comment)}
		}
	}

	return bare
}

func schemaTypeFromRef(ref string) string {
	u, err := url.Parse(ref)
	if err != nil {
		return path.Base(ref)
	}
	if typ := u.Query().Get("type"); typ != "" {
		return typ
	}
	if typ := u.Query().Get("source"); typ != "" {
		return typ
	}
	return path.Base(strings.TrimRight(u.Path, "/"))
}

func LookupSchema(schemaType string) (*jsonschema.Schema, error) {
	return schema.LookupSchemaType(schemaType)
}

// GetValidSchemaTypes returns all valid schema type names
func GetValidSchemaTypes() []string {
	return schema.GetSchemaTypes()
}

// IsValidSchemaType checks if a schema type is valid
func IsValidSchemaType(schemaType string) bool {
	return schema.IsValidSchemaType(schemaType)
}
