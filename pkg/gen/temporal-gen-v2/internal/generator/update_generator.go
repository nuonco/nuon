package generator

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"text/template"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/parser"
)

//go:embed templates/update.tmpl
var updateTemplateFS embed.FS

type UpdateData struct {
	ClientName   string
	Name         string
	OriginalName string
	UpdateName   string
	InputType    string
	OutputType   string
	Options      *parser.UpdateOptions
}

func GenerateUpdate(data UpdateData) ([]byte, error) {
	tmplContent, err := updateTemplateFS.ReadFile("templates/update.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("update").Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), fmt.Errorf("failed to format source: %w", err)
	}

	return formatted, nil
}
