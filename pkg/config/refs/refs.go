package refs

import (
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/mitchellh/reflectwalk"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
)

type Walker struct {
	refs []Ref
	fn   func(val string) []Ref
}

func (t *Walker) Struct(v reflect.Value) error {
	return nil
}

func (t *Walker) StructField(sf reflect.StructField, v reflect.Value) error {
	return nil
}

func (t *Walker) Array(v reflect.Value) error {
	return nil
}

func (t *Walker) ArrayElem(idx int, v reflect.Value) error {
	if v.Kind() == reflect.Struct {
		return t.Struct(v)
	}

	return t.Primitive(v)
}

func (t *Walker) Map(m reflect.Value) error {
	return nil
}

func (t *Walker) MapElem(m, k, v reflect.Value) error {
	return t.Primitive(v)
}

func (t *Walker) Primitive(v reflect.Value) error {
	var vals []Ref
	switch {
	case v.Kind() == reflect.String:
		vals = t.fn(v.String())
	case v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Uint8:
		vals = t.fn(string(v.Bytes()))
	}

	t.refs = append(t.refs, vals...)
	return nil
}

func Parse(obj any) ([]Ref, error) {
	walker := &Walker{
		refs: make([]Ref, 0),
		fn:   ParseFieldRefs,
	}

	if err := reflectwalk.Walk(obj, walker); err != nil {
		return nil, errors.Wrap(err, "unable to walk type for all inputs")
	}

	result := uniqueifyRefs(walker.refs)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Type != result[j].Type {
			return result[i].Type < result[j].Type
		}
		if result[i].Name != result[j].Name {
			return result[i].Name < result[j].Name
		}
		return result[i].Input < result[j].Input
	})
	return result, nil
}

// NOTE(jm): this was the fastest way to build out a list of all references, however long term we would like to switch
// to use an AST to identify all references in a "smarter" and less-brittle way.
//
// https://pkg.go.dev/text/template/parse
func ParseFieldRefs(inputVar string) []Ref {
	// Order matters: more-specific patterns (workflows, legacy install.* paths) run first.
	orderedPatterns := []struct {
		refType RefType
		pattern string
	}{
		{RefTypeActions, `nuon\.actions\.workflows\.([^.}\s]+)`},
		{RefTypeComponents, `nuon\.install\.components\.([^.}]+)\.outputs\.([^}\s]+)`},
		{RefTypeSandbox, `nuon\.install\.sandbox\.outputs\.([^}\s]+)`},
		{RefTypeComponents, `nuon\.components\.([^.}]+)\.outputs\.([^}\s]+)`},
		{RefTypeActions, `nuon\.actions\.([^.}]+)\.outputs\.([^.}\s]+)`},
		{RefTypeSecrets, `nuon\.secrets\.([^.}\s]+)`},
		{RefTypeInputs, `nuon\.inputs\.inputs\.([^.}\s]+)`},
		{RefTypeInstallInputs, `nuon\.install\.inputs\.([^.}\s]+)`},
		{RefTypeInstallStack, `nuon\.install_stack\.outputs\.([^.}\s]+)`},
		{RefTypeSandbox, `nuon\.sandbox\.outputs\.([^.}\s]+)`},
	}

	refs := make([]Ref, 0)
	for _, entry := range orderedPatterns {
		refType := entry.refType
		pattern := entry.pattern
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(inputVar, -1)

		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			r := Ref{
				Type:  refType,
				Name:  strings.TrimSpace(match[1]),
				Input: strings.TrimSpace(match[0]),
			}

			if generics.SliceContains(refType, []RefType{
				RefTypeActions,
				RefTypeComponents,
			}) {
				if len(match) >= 3 {
					r.Value = strings.TrimSpace(match[2])
				}
			}

			refs = append(refs, r)
		}
	}

	return refs
}
