package render

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
)

// RenderWithWarnings walks through the template variable by variable and will return well formed errors for any
// partial that was "unrenderable".
func RenderWithWarnings(inputVal string, data map[string]interface{}) (string, []error, error) {
	if inputVal == "" {
		return "", nil, nil
	}

	data = EnsurePrefix(data)

	vars := Parse(inputVal)
	warnings := make([]error, 0)
	for _, v := range vars {
		_, err := RenderVar(v, data)
		if err != nil {
			warnings = append(warnings, err)
			continue
		}
	}

	var err error
	if len(warnings) < 1 {
		inputVal, err = renderFinal(inputVal, data)
		if err != nil {
			return inputVal, []error{errors.Wrap(err, "unable to render template")}, nil
		}
	}

	return inputVal, warnings, nil
}

func renderFinal(inputVal string, data map[string]interface{}) (string, error) {
	funcMap := template.FuncMap{
		"now": time.Now,
	}

	temp, err := template.New("input").
		Funcs(funcMap).
		Funcs(sprig.FuncMap()).
		Option("missingkey=zero").
		Parse(inputVal)
	if err != nil {
		return inputVal, err
	}

	buf := new(bytes.Buffer)
	if err := temp.Execute(buf, data); err != nil {
		return inputVal, fmt.Errorf("unable to execute template: %w", err)
	}

	return buf.String(), nil
}

func RenderVar(v Var, data map[string]interface{}) (string, error) {
	temp, err := template.New("input").Option("missingkey=error").Parse(v.Template)
	if err != nil {
		return "", RenderErr{
			Template: v.Template,
			Name:     v.Name,
			Err:      errors.Wrap(err, "invalid template"),
		}
	}

	buf := new(bytes.Buffer)
	if err := temp.Execute(buf, data); err != nil {
		var execErr template.Error
		if errors.As(err, &execErr) {
			return "", RenderErr{
				Template: v.Template,
				Name:     v.Name,
				Err:      err,
			}
		}

		return "", fmt.Errorf("unable to execute template: %s", v.Template)
	}

	return buf.String(), nil
}
