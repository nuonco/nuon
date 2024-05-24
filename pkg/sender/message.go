package sender

import (
	"bytes"
	"fmt"
	"html/template"
)

func TemplateMessage(msg string, vars map[string]string) (string, error) {
	temp, err := template.New("msg").Option("missingkey=zero").Parse(msg)
	if err != nil {
		return "", fmt.Errorf("unable to create template: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := temp.Execute(buf, vars); err != nil {
		return "", fmt.Errorf("unable to execute template: %w", err)
	}

	return buf.String(), nil
}
