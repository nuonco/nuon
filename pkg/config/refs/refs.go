package refs

import (
	"reflect"
	"regexp"

	"github.com/mitchellh/reflectwalk"
	"github.com/pkg/errors"
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

	return uniqueifyRefs(walker.refs), nil
}

func ParseFieldRefs(inputVar string) []Ref {
	refPatterns := map[RefType]string{
		RefTypeInputs:     `nuon\.inputs\.([^.}]+)`,
		RefTypeComponents: `nuon\.components\.([^.}]+)\.outputs`,
		// RefTypeComponentsNested: `nuon\.components\.components\.([^.]+)\.outputs`,
		RefTypeInstallStack: `nuon\.install_stack\.outputs\.([^.}]+)`,
		RefTypeSandbox:      `nuon\.sandbox\.outputs\.([^.}]+)`,
	}

	refs := make([]Ref, 0)
	for refType, pattern := range refPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(inputVar, -1)

		for _, match := range matches {
			if len(match) > 1 {
				refs = append(refs, Ref{
					Type: refType,
					Name: match[1],
				})
			}
		}
	}

	return refs
}
