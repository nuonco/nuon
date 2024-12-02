package main

import (
	"go/ast"
	"testing"
)

func TestTitleize(t *testing.T) {
	var tt = map[string]struct {
		in, out *ast.Ident
	}{
		"simple": {
			in:  &ast.Ident{Name: "foo"},
			out: &ast.Ident{Name: "Foo"},
		},
		"underscore": {
			in:  &ast.Ident{Name: "foo_bar"},
			out: &ast.Ident{Name: "FooBar"},
		},
		"allcaps": {
			in:  &ast.Ident{Name: "FOO_BAR"},
			out: &ast.Ident{Name: "FOOBAR"},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			if got := titleize(tc.in); got.Name != tc.out.Name {
				t.Errorf("expected %q, got %q", tc.out.Name, got.Name)
			}
		})
	}
}
