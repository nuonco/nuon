package main

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"regexp"
	"strings"
)

var goDirective = regexp.MustCompile(`^//(go:|line |export |extern |nolint|lint:|revive:|#nosec|sumtype:)|^//\s*(nolint|#nosec|nosec|\+build|\+[a-z]|swagger:)`)

var annotationLine = regexp.MustCompile(`^@[A-Za-z]`)

func checkGo(path string, src []byte, o options) ([]violation, []byte, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	keep := map[*ast.CommentGroup]bool{}
	keepAll := func(groups ...*ast.CommentGroup) {
		for _, g := range groups {
			if g != nil {
				keep[g] = true
			}
		}
	}
	keepWithin := func(n ast.Node) {
		for _, g := range f.Comments {
			if g.Pos() >= n.Pos() && g.End() <= n.End() {
				keep[g] = true
			}
		}
	}

	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			if strings.HasPrefix(d.Name.Name, "Example") && d.Body != nil {
				for _, g := range f.Comments {
					if g.Pos() < d.Body.Pos() || g.End() > d.Body.End() {
						continue
					}
					t := strings.TrimSpace(g.Text())
					if strings.HasPrefix(t, "Output:") || strings.HasPrefix(t, "Unordered output:") {
						keep[g] = true
					}
				}
			}
		case *ast.GenDecl:
			switch d.Tok {
			case token.IMPORT:
				for _, s := range d.Specs {
					if is, ok := s.(*ast.ImportSpec); ok && is.Path.Value == `"C"` {
						keepAll(d.Doc, is.Doc)
					}
				}
			case token.TYPE:
				for _, s := range d.Specs {
					ts := s.(*ast.TypeSpec)
					if !o.apiType(f.Name.Name, ts.Name.Name) {
						continue
					}
					keepAll(ts.Doc, ts.Comment)
					if len(d.Specs) == 1 {
						keepAll(d.Doc)
					}
					ast.Inspect(ts.Type, func(n ast.Node) bool {
						if fld, ok := n.(*ast.Field); ok {
							keepAll(fld.Doc, fld.Comment)
						}
						return true
					})
				}
			case token.CONST:
				for _, s := range d.Specs {
					vs := s.(*ast.ValueSpec)
					if id, ok := vs.Type.(*ast.Ident); ok && o.apiType(f.Name.Name, id.Name) {
						keepWithin(d)
						break
					}
				}
			}
		}
	}

	var vs []violation
	for _, g := range f.Comments {
		if keep[g] || keepGroup(g) {
			continue
		}
		deprecatedFrom := -1
		for i, c := range g.List {
			if strings.HasPrefix(commentText(c.Text), "Deprecated:") {
				deprecatedFrom = i
				break
			}
		}
		for i, c := range g.List {
			if goDirective.MatchString(c.Text) || (deprecatedFrom >= 0 && i >= deprecatedFrom) {
				continue
			}
			start := fset.Position(c.Pos()).Offset
			vs = append(vs, violation{
				line:  fset.Position(c.Pos()).Line,
				text:  c.Text,
				start: start,
				end:   start + len(c.Text),
			})
		}
	}
	if len(vs) == 0 {
		return nil, src, nil
	}

	fixed, err := format.Source(removeComments(src, vs, false))
	if err != nil {
		return vs, src, nil
	}
	return vs, fixed, nil
}

func (o options) apiType(pkg, name string) bool {
	if o.swaggerTypes == nil {
		return true
	}
	return o.swaggerTypes[pkg+"."+name]
}

func keepGroup(g *ast.CommentGroup) bool {
	if len(g.List) == 0 {
		return true
	}
	if keepFirstLine(commentText(g.List[0].Text)) {
		return true
	}
	for _, c := range g.List {
		t := commentText(c.Text)
		if annotationLine.MatchString(t) || license(t) {
			return true
		}
	}
	return false
}

func commentText(raw string) string {
	t := strings.TrimPrefix(raw, "//")
	t = strings.TrimPrefix(t, "/*")
	t = strings.TrimSuffix(t, "*/")
	return strings.TrimSpace(t)
}

var keepPrefixes = []string{"why:", "TODO", "FIXME", "HACK", "XXX"}

func keepFirstLine(t string) bool {
	for _, p := range keepPrefixes {
		if strings.HasPrefix(t, p) {
			return true
		}
	}
	return false
}

func license(t string) bool {
	return strings.Contains(t, "Copyright") || strings.Contains(t, "SPDX-License-Identifier") ||
		strings.Contains(t, "Licensed under")
}
