package generator

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"text/template"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/parser"
)

//go:embed templates/signal.tmpl
var signalTemplateFS embed.FS

type SignalData struct {
	ClientName   string
	Name         string
	OriginalName string
	InputType    string
	Options      *parser.SignalOptions
}

func GenerateSignal(data SignalData) ([]byte, error) {
	tmplContent, err := signalTemplateFS.ReadFile("templates/signal.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("signal").Parse(string(tmplContent))
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
