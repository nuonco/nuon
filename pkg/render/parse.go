package render

import (
	"regexp"
	"strings"
)

type Var struct {
	Template string
	Name     string
}

func Parse(str string) []Var {
	re := regexp.MustCompile(`\{\{(.*?)\}\}`)
	matches := re.FindAllStringSubmatch(str, -1)

	vars := make([]Var, 0)
	lookup := make(map[string]struct{}, 0)

	for _, matchP := range matches {
		tmpl := matchP[0]

		name := matchP[0]
		name = strings.ReplaceAll(name, "{{", "")
		name = strings.ReplaceAll(name, "}}", "")
		name = strings.TrimSpace(name)
		name = strings.Replace(name, ".", "", 1)

		if !strings.HasPrefix(name, defaultPrefix) {
			continue
		}

		if name == "" {
			continue
		}

		if _, found := lookup[tmpl]; found {
			continue
		}

		lookup[tmpl] = struct{}{}
		vars = append(vars, Var{
			Template: matchP[0],
			Name:     name,
		})
	}

	return vars
}
