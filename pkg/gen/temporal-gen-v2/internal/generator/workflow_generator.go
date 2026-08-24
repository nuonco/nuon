package generator

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"text/template"
	"time"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/parser"
)

//go:embed templates/workflow.tmpl
var workflowTemplateFS embed.FS

type WorkflowData struct {
	Name         string
	OriginalName string
	InputType    string
	OutputType   string
	Options      *parser.WorkflowOptions
	Receiver     string
}

func GenerateWorkflow(data WorkflowData) ([]byte, error) {
	tmplContent, err := workflowTemplateFS.ReadFile("templates/workflow.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("workflow").Funcs(template.FuncMap{
		"durationToNs": func(d time.Duration) int64 {
			return d.Nanoseconds()
		},
	}).Parse(string(tmplContent))
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
		// Return unformatted code for debugging
		return buf.Bytes(), fmt.Errorf("failed to format source: %w", err)
	}

	return formatted, nil
}
