package generator

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/parser"
)

//go:embed templates/activity.tmpl templates/activity_stub.tmpl templates/activity_registration.tmpl
var activityTemplateFS embed.FS

type Param struct {
	Name         string
	Type         string
	ExportedName string
}

type ActivityData struct {
	Name          string
	OriginalName  string
	QualifiedName string
	InputType     string
	OutputType    string
	Options       *parser.ActivityOptions
	Params        []Param
	Receiver      string
	ByFieldType   string
}

func GenerateActivity(data ActivityData) ([]byte, error) {
	return generateActivityWithMode(data, false)
}

func GenerateActivityStub(data ActivityData) ([]byte, error) {
	// For stubs, generate minimal valid code without templates
	var buf bytes.Buffer

	// Generate request type if needed
	if data.Options.GenerateWrapper && len(data.Params) > 0 {
		reqType := toPascalCase(data.Name) + "Request"
		buf.WriteString(fmt.Sprintf("type %s struct {\n", reqType))
		for _, p := range data.Params {
			buf.WriteString(fmt.Sprintf("\t%s interface{}\n", p.ExportedName))
		}
		buf.WriteString("}\n\n")

		// Generate wrapper method
		buf.WriteString(fmt.Sprintf("func (a %s) ", data.Receiver))
		if data.Options.WrapperPrefix != "" {
			buf.WriteString(data.Options.WrapperPrefix)
		}
		buf.WriteString(toPascalCase(data.Name))
		buf.WriteString(fmt.Sprintf("(ctx context.Context, req %s) ", reqType))
		if data.OutputType != "" && data.OutputType != "interface{}" {
			buf.WriteString("(interface{}, error) {\n")
		} else {
			buf.WriteString("error {\n")
		}
		buf.WriteString("\tpanic(\"stub implementation - will be replaced in phase 2\")\n")
		buf.WriteString("}\n\n")
	}

	// Generate Await function
	buf.WriteString(fmt.Sprintf("func Await%s(ctx workflow.Context, input interface{}, opts ...*workflow.ActivityOptions) ", toPascalCase(data.Name)))
	if data.OutputType != "" && data.OutputType != "interface{}" {
		buf.WriteString("(interface{}, error) {\n")
	} else {
		buf.WriteString("error {\n")
	}
	buf.WriteString("\tpanic(\"stub implementation - will be replaced in phase 2\")\n")
	buf.WriteString("}\n\n")

	// Generate ByField function if needed
	if data.Options.ByField != "" && !data.Options.ByFieldOnly {
		buf.WriteString(fmt.Sprintf("func Await%sBy%s(ctx workflow.Context, input interface{}, opts ...*workflow.ActivityOptions) ", toPascalCase(data.Name), data.Options.ByField))
		if data.OutputType != "" && data.OutputType != "interface{}" {
			buf.WriteString("(interface{}, error) {\n")
		} else {
			buf.WriteString("error {\n")
		}
		buf.WriteString("\tpanic(\"stub implementation - will be replaced in phase 2\")\n")
		buf.WriteString("}\n\n")
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), fmt.Errorf("failed to format stub source: %w", err)
	}

	return formatted, nil
}

func generateActivityWithMode(data ActivityData, stubMode bool) ([]byte, error) {
	templateFile := "templates/activity.tmpl"
	if stubMode {
		templateFile = "templates/activity_stub.tmpl"
	}

	tmplContent, err := activityTemplateFS.ReadFile(templateFile)
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
		"isPointer": func(s string) bool {
			return strings.HasPrefix(s, "*")
		},
		"derefType": func(s string) string {
			return strings.TrimPrefix(s, "*")
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
		// Save unformatted code for debugging
		debugPath := fmt.Sprintf("/tmp/temporal-gen-debug-%s.go", data.Name)
		if debugErr := os.WriteFile(debugPath, buf.Bytes(), 0644); debugErr == nil {
			return buf.Bytes(), fmt.Errorf("failed to format source (saved to %s for debugging): %w", debugPath, err)
		}
		// Return unformatted code for debugging if formatting fails
		return buf.Bytes(), fmt.Errorf("failed to format source: %w", err)
	}

	return formatted, nil
}

func GenerateActivityRegistration(data ActivityData) ([]byte, error) {
	tmplContent, err := activityTemplateFS.ReadFile("templates/activity_registration.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read registration template: %w", err)
	}

	tmpl, err := template.New("activity_registration").Funcs(template.FuncMap{
		"ToPascal": func(s string) string {
			if s == "" {
				return ""
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
	}).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse registration template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute registration template: %w", err)
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Return unformatted code for debugging if formatting fails
		return buf.Bytes(), fmt.Errorf("failed to format registration source: %w", err)
	}

	return formatted, nil
}
