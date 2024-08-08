package main

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	GenMarker = "@await-gen"
)

func loadBase(ctx context.Context, dir string) ([]*BaseFile, error) {
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Error parsing package:", err)
		os.Exit(1)
	}

	var walkerr error
	var ret []*BaseFile
	// Inspect each package
	for _, pkg := range pkgs {
		// Inspect each file in the package
		for fpath, file := range pkg.Files {
			var bfs []BaseFn
			if walkerr != nil {
				return nil, walkerr
			}
			// Walk the AST
			ast.Inspect(file, func(n ast.Node) bool {
				if walkerr != nil {
					return false
				}
				switch x := n.(type) {
				case *ast.FuncDecl:
					if x.Doc == nil {
						return true
					}

					opts, err := extractGenOptions(fset, x.Doc)
					if err != nil {
						walkerr = err
						return false
					}
					if opts != nil {
						if len(x.Type.Params.List) != 2 {
							walkerr = withPos(fset, x.Type.Params.Pos(), errors.New("base activity func must have exactly two params, ctx and a request object"))
							return false
						}

						bfs = append(bfs, BaseFn{
							Fn:   x,
							Opts: opts,
						})
					}
				}
				return true
			})

			if len(bfs) > 0 {
				ret = append(ret, &BaseFile{
					Path: fpath,
					File: file,
					Fns:  bfs,
				})
			}
		}
	}

	if walkerr != nil {
		return nil, walkerr
	}

	if err != nil {
		// TODO remember how to get the module/pkg path, not just filepath
		fmt.Fprintf(os.Stderr, "failed to load base package %q: %s", dir, err)
		os.Exit(1)
	}

	if len(pkgs) != 1 {
		fmt.Fprintf(os.Stderr, "expected there to be exactly one package in directory %q, got %d", dir, len(pkgs))
		os.Exit(1)
	}

	return ret, nil
}

func extractGenOptions(fset *token.FileSet, cg *ast.CommentGroup) (*GenOptions, error) {
	if cg == nil {
		return nil, nil
	}
	ret := new(GenOptions)
	var matched bool
	for _, com := range cg.List {
		parts := strings.Split(com.Text, " ")

		// TODO(sdboyer) more validation
		switch len(parts) {
		case 0, 1:
			continue
		case 2:
			if parts[1] == GenMarker {
				if matched {
					return nil, withPos(fset, com.Pos(), fmt.Errorf("%s may be declared only once", GenMarker))
				}
				matched = true
			}
		case 3:
			switch parts[1] {
			case "@execution-timeout":
				var err error
				ret.Timeout, err = time.ParseDuration(parts[2])
				if err != nil {
					return nil, withPos(fset, com.Pos(), fmt.Errorf("@execution-timeout must be a valid Go duration string per https://pkg.go.dev/time#ParseDuration, got %q", parts[2]))
				}
			case "@max-retries":
				var err error
				ret.MaxRetries, err = strconv.Atoi(parts[2])
				if err != nil {
					return nil, withPos(fset, com.Pos(), fmt.Errorf("@max-retries must be a valid Go duration string, got %q", parts[2]))
				}
			}
		}
	}

	if matched {
		return ret, nil
	}
	return nil, nil
}
