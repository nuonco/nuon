package main

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/dave/dst/decorator"
	"github.com/grafana/codejen"
	"golang.org/x/tools/imports"
)

func GoImportsMapper(f codejen.File) (codejen.File, error) {
	buf := new(bytes.Buffer)

	fset := token.NewFileSet()
	fname := filepath.Base(f.RelativePath)
	gf, err := decorator.ParseFile(fset, fname, f.Data, parser.ParseComments)
	if err != nil {
		return codejen.File{}, fmt.Errorf("error parsing generated file: %w", err)
	}

	err = decorator.Fprint(buf, gf)
	if err != nil {
		return codejen.File{}, fmt.Errorf("error formatting generated file: %w", err)
	}

	byt, err := imports.Process(fname, buf.Bytes(), nil)
	if err != nil {
		return codejen.File{}, fmt.Errorf("goimports processing of generated file failed: %w", err)
	}

	// Compare imports before and after; warn about performance if some were added
	gfa, _ := parser.ParseFile(fset, fname, string(byt), parser.ParseComments)
	imap := make(map[string]bool)
	for _, im := range gf.Imports {
		imap[im.Path.Value] = true
	}
	var added []string
	for _, im := range gfa.Imports {
		if !imap[im.Path.Value] {
			added = append(added, im.Path.Value)
		}
	}

	if len(added) != 0 {
		// TODO improve the guidance in this error if/when we better abstract over imports to generate
		return codejen.File{}, fmt.Errorf("goimports added the following import statements to %s: \n\t%s\nRelying on goimports to find imports significantly slows down code generation. Either add these imports with an AST manipulation in cfg.ApplyFuncs, or set cfg.IgnoreDiscoveredImports to true", f.RelativePath, strings.Join(added, "\n\t"))
	}
	f.Data = byt
	return f, nil
}
