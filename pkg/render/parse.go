package render

import (
	"regexp"
	"strings"
)

type Var struct {
	Template string
	Name     string
}

// Parse returns a list of all the template uses in a string, which allows for better rendering in the case of errors.
// NOTE: this can _not_ be used to reconstruct the original template using strings.Join() as any actual content will be
// removed.
func Parse(str string) []Var {
	re := regexp.MustCompile(`\{\{(.*?)\}\}`)
	matches := re.FindAllStringSubmatch(str, -1)

	vars := make([]Var, 0)
	lookup := make(map[string]struct{}, 0)

	for _, matchP := range matches {
		// get the current match, minus go strings, for the current depth
		tmpl := matchP[0]

		name := matchP[0]
		name = strings.ReplaceAll(name, "{{", "")
		name = strings.ReplaceAll(name, "}}", "")
		name = strings.TrimSpace(name)
		name = strings.Replace(name, ".", "", 1)

		// any value that is not a prefix, we do not validate, as this could be a var or other
		if !strings.HasPrefix(name, defaultPrefix) {
			continue
		}

		// this could only happen if a user added `{{}}`
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
