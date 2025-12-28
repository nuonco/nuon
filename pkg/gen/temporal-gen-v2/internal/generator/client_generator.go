package generator

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"text/template"
)

//go:embed templates/client.tmpl
var clientTemplateFS embed.FS

type ClientData struct {
	ClientName string
}

func GenerateClient(data ClientData) ([]byte, error) {
	tmplContent, err := clientTemplateFS.ReadFile("templates/client.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("client").Parse(string(tmplContent))
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
