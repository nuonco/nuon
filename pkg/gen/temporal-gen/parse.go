package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/go-toolsmith/astfmt"
	"golang.org/x/tools/go/packages"
)

const (
	GenMarker = "@temporal-gen"
)

func parseDir(ctx context.Context, dir string) ([]*BaseFile, error) {
	fset := token.NewFileSet()

	pkgs, err := packages.Load(&packages.Config{
		Fset:    fset,
		Context: ctx,
		Mode:    packages.NeedName | packages.NeedCompiledGoFiles | packages.NeedFiles | packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
	}, dir)
	if err != nil {
		fmt.Println("Error parsing package:", err)
		os.Exit(1)
	}

	var walkerr error
	var ret []*BaseFile
	// Inspect each package
	for _, pkg := range pkgs {
		// Inspect each file in the package
		for i, file := range pkg.Syntax {
			// TODO(sdboyer): this filename may be wrong if syntax checking returned nil for any files in the pkg
			fpath := filepath.Base(pkg.CompiledGoFiles[i])
			var actfns []ActivityFn
			var asactfns []AsActivityFn
			var wkffns []WorkflowFn
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

					for _, com := range x.Doc.List {
						parts := strings.Split(com.Text, " ")

						// TODO(sdboyer) more validation
						switch len(parts) {
						case 0, 1, 2:
							continue
						case 3:
							if parts[1] == GenMarker {
								switch parts[2] {
								// TODO(sdboyer) enforce that first param is context
								case "as-activity":
									if len(x.Type.Params.List) < 2 {
										fmt.Println(astfmt.Sprint(x.Type.Params.List))
										walkerr = withPos(fset, x.Type.Params.Pos(), errors.New("base func must have at least two params, the first being ctx"))
										return false
									}
									afn, err := extractAsActivityFn(fset, x, pkg)
									if err != nil {
										walkerr = err
										return false
									}
									if afn != nil {
										asactfns = append(asactfns, *afn)
									}

								case "activity", "workflow":
									if len(x.Type.Params.List) != 2 {
										walkerr = withPos(fset, x.Type.Params.Pos(), errors.New("base func must have exactly two params, ctx and a request object"))
										return false
									}
									switch parts[2] {
									case "activity":
										afn, err := extractActivityFn(fset, x, pkg)
										if err != nil {
											walkerr = err
											return false
										}
										if afn != nil {
											actfns = append(actfns, *afn)
										}
									case "workflow":
										wfn, err := extractWorkflowFn(fset, x, pkg)
										if err != nil {
											walkerr = err
											return false
										}
										if wfn != nil {
											wkffns = append(wkffns, *wfn)
										}
									}
								default:
									walkerr = withPos(fset, x.Type.Params.Pos(), errors.Newf("unrecognized temporal-gen directive %q", parts[2]))
								}
							}
						}
					}
				}
				return true
			})

			if len(actfns) > 0 || len(wkffns) > 0 || len(asactfns) > 0 {
				ret = append(ret, &BaseFile{
					Path:          fpath,
					File:          file,
					ActivityFns:   actfns,
					WorkflowFns:   wkffns,
					AsActivityFns: asactfns,
					Package:       pkg,
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

func extractActivityFn(fset *token.FileSet, fn *ast.FuncDecl, pkg *packages.Package) (*ActivityFn, error) {
	cg := fn.Doc

	ret := new(ActivityGenOptions)
	for _, com := range cg.List {
		parts := strings.Split(com.Text, " ")

		// TODO(sdboyer) more validation
		switch len(parts) {
		case 0, 1:
			continue
		case 3:
			switch parts[1] {
			case "@schedule-to-close-timeout":
				var err error
				ret.ScheduleToCloseTimeout, err = time.ParseDuration(parts[2])
				if err != nil {
					return nil, withPos(fset, com.Pos(), fmt.Errorf("@execution-timeout must be a valid Go duration string per https://pkg.go.dev/time#ParseDuration, got %q", parts[2]))
				}
			case "@start-to-close-timeout":
				var err error
				ret.StartToCloseTimeout, err = time.ParseDuration(parts[2])
				if err != nil {
					return nil, withPos(fset, com.Pos(), fmt.Errorf("@start-to-close-timeout must be a valid Go duration string per https://pkg.go.dev/time#ParseDuration, got %q", parts[2]))
				}
			case "@max-retries":
				var err error
				ret.MaxRetries, err = strconv.Atoi(parts[2])
				if err != nil {
					return nil, withPos(fset, com.Pos(), fmt.Errorf("@max-retries must be a valid Go duration string, got %q", parts[2]))
				}
			case "@options-callback":
				ret.OptionsCallback = parts[2]
			case "@by-id":
				var reqt *types.Struct
				var ok bool
				reqtype := fn.Type.Params.List[1].Type
				if reqti, has := pkg.TypesInfo.Types[fn.Type.Params.List[1].Type]; !has {
					return nil, withPos(fset, com.Pos(), fmt.Errorf("internal error - no type info could be found for %s", astfmt.Sprint(reqtype)))
				} else {
					rtyp := reqti.Type
					if ptr, is := rtyp.(*types.Pointer); is {
						rtyp = ptr.Elem()
					}
					if reqt, ok = rtyp.Underlying().(*types.Struct); !ok {
						return nil, withPos(fset, com.Pos(), fmt.Errorf("@by-id can only be used when the function's second parameter is struct-kinded, but %s is not", astfmt.Sprint(reqtype)))
					}
				}

				var match *types.Var
				for i := 0; i < reqt.NumFields(); i++ {
					if reqt.Field(i).Name() == parts[2] {
						match = reqt.Field(i)
					}
				}
				if match == nil {
					return nil, withPos(fset, com.Pos(), fmt.Errorf("@by-id must be provided the name of a field on %s; got %q", astfmt.Sprint(reqtype), parts[2]))
				}
				ret.ById = ByIdOptions{
					Name: match.Name(),
					Type: match.Type().Underlying().String(),
				}
			case "@by-id-only":
				ret.ByIdOnly = true
			}
		}
	}

	if ret.ByIdOnly && ret.ById.Name == "" {
		return nil, withPos(fset, fn.Pos(), errors.New("@by-id-only may not be specified without also providing a field name via @by-id"))
	}

	return &ActivityFn{
		Fn:   fn,
		Opts: ret,
	}, nil
}

func extractAsActivityFn(fset *token.FileSet, fn *ast.FuncDecl, pkg *packages.Package) (*AsActivityFn, error) {
	inner, err := extractActivityFn(fset, fn, pkg)
	if err != nil {
		return nil, err
	}

	// by-id not directly allowed on as-activity
	if inner.Opts.ById.Name != "" {
		for _, com := range fn.Doc.List {
			parts := strings.Split(com.Text, " ")
			if len(parts) == 3 && parts[1] == "@by-id" {
				return nil, withPos(fset, com.Pos(), errors.New("@by-id may not be specified on as-activity functions, they are created automatically where possible"))
			}
		}
	}

	return &AsActivityFn{
		Fn: fn,
		Opts: &AsActivityGenOptions{
			Inner: inner.Opts,
		},
	}, nil
}

func extractWorkflowFn(fset *token.FileSet, fn *ast.FuncDecl, pkg *packages.Package) (*WorkflowFn, error) {
	cg := fn.Doc

	ret := new(WorkflowGenOptions)
	for _, com := range cg.List {
		parts := strings.Split(com.Text, " ")

		// TODO(sdboyer) more validation
		switch len(parts) {
		case 0, 1:
			continue
		case 3:
			switch parts[1] {
			case "@execution-timeout":
				var err error
				ret.ExecutionTimeout, err = time.ParseDuration(parts[2])
				if err != nil {
					return nil, withPos(fset, com.Pos(), fmt.Errorf("@execution-timeout must be a valid Go duration string per https://pkg.go.dev/time#ParseDuration, got %q", parts[2]))
				}
			case "@task-timeout":
				var err error
				ret.TaskTimeout, err = time.ParseDuration(parts[2])
				if err != nil {
					return nil, withPos(fset, com.Pos(), fmt.Errorf("@task-timeout must be a valid Go duration string per https://pkg.go.dev/time#ParseDuration, got %q", parts[2]))
				}
			case "@id-callback":
				ret.IDCallback = parts[2]
			case "@id-template":
				ret.IDTemplate = strings.Join(parts[2:], " ")
				if _, err := template.New("workflowID").Parse(ret.IDTemplate); err != nil {
					return nil, withPos(fset, com.Pos(), fmt.Errorf("@id-template must be a valid Go template string: %s", err))
				}
			case "@wait-for-cancellation":
				var err error
				ret.WaitForCancellation, err = strconv.ParseBool(parts[2])
				if err != nil {
					return nil, withPos(fset, com.Pos(), fmt.Errorf("@wait-for-cancellation must be either 'true' or 'false', got %q", parts[2]))
				}
			case "@task-queue":
				ret.TaskQueue = parts[2]
			case "@options-callback":
				ret.OptionsCallback = parts[2]
			}
		default:
			switch parts[1] {
			case "@id-template":
				ret.IDTemplate = strings.Join(parts[2:], " ")
				if _, err := template.New("workflowID").Parse(ret.IDTemplate); err != nil {
					return nil, withPos(fset, com.Pos(), fmt.Errorf("@id-template must be a valid Go template string: %s", err))
				}
			}
		}
	}


	if ret.IDCallback != "" && ret.IDTemplate != "" {
		return nil, withPos(fset, fn.Pos(), errors.New("@id-callback and @id-template may not be specified together"))
	}

	return &WorkflowFn{
		Fn:   fn,
		Opts: ret,
	}, nil
}
