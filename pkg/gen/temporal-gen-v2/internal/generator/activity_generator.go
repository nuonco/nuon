package generator

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"strings"
	"text/template"
	"time"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/parser"
)

//go:embed templates/activity.tmpl
var activityTemplateFS embed.FS

type Param struct {
	Name         string
	Type         string
	ExportedName string
}

type ActivityData struct {
	Name         string
	OriginalName string
	InputType    string
	OutputType   string
	Options      *parser.ActivityOptions
	Params       []Param
	Receiver     string
	ByFieldType  string
}

func GenerateActivity(data ActivityData) ([]byte, error) {
	tmplContent, err := activityTemplateFS.ReadFile("templates/activity.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	tmpl, err := template.New("activity").Funcs(template.FuncMap{
		"durationToNs": func(d time.Duration) int64 {
			return d.Nanoseconds()
		},
		"ToPascal": func(s string) string {
			if s == "" {
				return ""
			}
			return strings.ToUpper(s[:1]) + s[1:]
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
		// Return unformatted code for debugging if formatting fails
		return buf.Bytes(), fmt.Errorf("failed to format source: %w", err)
	}

	return formatted, nil
}
