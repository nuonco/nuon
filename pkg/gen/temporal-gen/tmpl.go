package main

import (
	"embed"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/grafana/codejen"
)

//go:embed tmpl/*.tmpl
var tmplFS embed.FS

// All the parsed templates in the tmpl subdirectory
var tmpls *template.Template

func init() {
	base := template.New("codegen").Funcs(template.FuncMap{
		"TrimPtr":     func(s string) string { return strings.TrimPrefix(s, "*") },
		"DurationLit": durationToGoCode,
	}).Funcs(sprig.FuncMap())
	tmpls = template.Must(base.ParseFS(tmplFS, "tmpl/*.tmpl"))
}

type (
	tvars_gen_header struct {
		MainGenerator string
		Using         []codejen.NamedJenny
		From          string
		Leader        string
	}
)

func durationToGoCode(d time.Duration) string {
	if d == 0 {
		return "0"
	}

	var result string

	// Handle negative durations
	if d < 0 {
		result = "-"
		d = -d
	}

	units := []struct {
		duration time.Duration
		name     string
	}{
		{time.Hour, "time.Hour"},
		{time.Minute, "time.Minute"},
		{time.Second, "time.Second"},
		{time.Millisecond, "time.Millisecond"},
		{time.Microsecond, "time.Microsecond"},
		{time.Nanosecond, "time.Nanosecond"},
	}

	for _, unit := range units {
		if d >= unit.duration {
			value := d / unit.duration
			d %= unit.duration
			if result != "" {
				result += " + "
			}
			result += fmt.Sprintf("%d * %s", value, unit.name)
		}
	}

	return result
}
