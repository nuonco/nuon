package render

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/pkg/errors"
)

// RenderTextV2 does the same thing as RenderV2, but using "text/template" instead of "html/template", so special
// characters are not escaped. Use it for values that end up in infrastructure APIs rather than in a browser: an input
// value containing "&" or a quote renders as "&amp;" / "&#34;" through RenderV2, which silently corrupts things like
// CloudFormation parameters.
func RenderTextV2(inputVal string, data map[string]interface{}) (string, error) {
	data = EnsurePrefix(data)

	if !strings.Contains(inputVal, ".nuon") {
		return inputVal, nil
	}

	temp, err := newTextTemplate(inputVal)
	if err != nil {
		return inputVal, err
	}

	buf := new(bytes.Buffer)
	if err := temp.Execute(buf, data); err != nil {
		return inputVal, fmt.Errorf("unable to execute template: %w", err)
	}

	return buf.String(), nil
}

// ValidateTextTemplate reports whether inputVal parses against the same function set RenderTextV2 executes with. Use it
// to reject a bad template early (e.g. at config sync) instead of failing later at render time.
func ValidateTextTemplate(inputVal string) error {
	_, err := newTextTemplate(inputVal)
	return err
}

func newTextTemplate(inputVal string) (*template.Template, error) {
	funcMap := template.FuncMap{
		"now": time.Now,
	}

	return template.New("input").
		Funcs(funcMap).
		Funcs(sprig.TxtFuncMap()).
		Option("missingkey=error").
		Parse(inputVal)
}

// RenderTextStringMap renders every value of m in place using RenderTextV2.
func RenderTextStringMap(m map[string]string, data map[string]interface{}) error {
	for key, val := range m {
		rendered, err := RenderTextV2(val, data)
		if err != nil {
			return errors.Wrapf(err, "unable to render %q", key)
		}

		m[key] = rendered
	}

	return nil
}
