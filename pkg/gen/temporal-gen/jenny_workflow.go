package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"

	"github.com/go-toolsmith/astfmt"
	"github.com/grafana/codejen"
)

type tvars_workflowfn struct {
	// Options contains the input-specified options that specify ChildWorkflowOptions and govern
	// various other generator behaviors
	Options WorkflowGenOptions

	// FnName is the name to use for the generated await function.
	FnName string

	// BaseFnSymbol is a string literal containing the name of the underlying activity function to be called. It includes a receiver if the function is being generated as a method
	BaseFnSymbol string

	// BaseFnName is a doc-friendly qualified name of the base function being referenced. If a function, it will just be the function name. If a method, it will be <ReceiverType>.<FuncName>
	BaseFnName string

	// ReqType is the naem of the request object type as a string literal
	ReqType string

	// IsMethod indicates whether the function should be generated as a method on the same type as the
	// base function it wraps
	IsMethod bool

	// RecvType is the var declaration and name of the receiver type as a string literal. Populated only if IsMethod is true
	Receiver string

	// HasResp indicates whether there is a non-error return value from the function
	HasResp bool

	// RespType is a string literal representing the qualified type of the response
	RespType string

	// RespIsPtr indicates whether the type of the return value is a pointer
	RespIsPtr bool

	// ReqIsPtr indicates whether the type of the request value is a pointer
	ReqIsPtr bool

	// Zero is a string literal representing the zero value of the response type
	Zero string
}

// WorkflowJenny is a jenny that generates a function that calls a provided base function as a Temporal workflow, and await the result.
type WorkflowJenny struct {
	UseMethods bool
}

func (j WorkflowJenny) JennyName() string {
	return "WorkflowJenny"
}

func (j WorkflowJenny) Generate(bf *BaseFile) (*codejen.File, error) {
	if len(bf.WorkflowFns) == 0 {
		return nil, nil
	}

	var buf bytes.Buffer
	genImports(&buf, bf)
	for _, bfn := range bf.WorkflowFns {
		tvars := tvars_workflowfn{
			Options: *bfn.Opts,
		}

		bfname := bfn.Fn.Name.String()
		tvars.FnName = fmt.Sprintf("Await%s", strings.Title(bfname))

		tvars.ReqType = astfmt.Sprint(bfn.Fn.Type.Params.List[1].Type)
		// _, tvars.ReqIsPtr = bfn.Fn.Type.Params.List[1].Type.(*ast.StarExpr)

		basebuf := new(strings.Builder)
		if bfn.Fn.Recv != nil {
			basebuf.WriteString("(")
			expr := bfn.Fn.Recv.List[0].Type
			if x, isPtr := expr.(*ast.StarExpr); isPtr {
				expr = x.X
				basebuf.WriteString("&")
			}
			// Injecting {} hardcodes assumption that receiver is struct-kinded
			basebuf.WriteString(astfmt.Sprint(expr) + "{}).")
			tvars.IsMethod = j.UseMethods
			recvVar := bfn.Fn.Recv.List[0].Names[0].Name
			tvars.Receiver = fmt.Sprintf("%s %s", recvVar, astfmt.Sprint(bfn.Fn.Recv.List[0].Type))

			tvars.BaseFnName = fmt.Sprintf("%s.%s", astfmt.Sprint(bfn.Fn.Recv.List[0].Type), bfname)
		} else {
			tvars.BaseFnName = bfname
		}
		basebuf.WriteString(bfname)

		tvars.BaseFnSymbol = basebuf.String()

		switch len(bfn.Fn.Type.Results.List) {
		case 1:
			// This template assumes without verification that the wrapped fn has a bare error return, b/c temporal requires that
			err := tmpls.Lookup("workflow_fn.tmpl").Execute(&buf, tvars)
			if err != nil {
				return nil, fmt.Errorf("error executing template for fn %s: %w", bfn.Fn.Name.String(), err)
			}
		case 2:
			// This template assumes without verification that the wrapped fn is a (custom value, error) return, b/c temporal requires that
			tvars.HasResp = true
			respt := bfn.Fn.Type.Results.List[0].Type
			tvars.RespType = astfmt.Sprint(respt)
			_, tvars.RespIsPtr = bfn.Fn.Type.Results.List[0].Type.(*ast.StarExpr)

			if !tvars.RespIsPtr {
				// Non-pointer return types need to be zero-initialized, the syntax for which
				// varies by type
				rt := bf.Package.TypesInfo.Types[respt].Type
				tvars.Zero = zerostr(rt)

				// If the return type is a struct imported from a different package, we have to qualify it
				if x, ok := rt.(*types.Named); ok && x.Obj().Pkg().Path() != bf.Package.PkgPath {
					if _, is := x.Underlying().(*types.Struct); is {
						tvars.Zero = fmt.Sprintf("%s.%s", x.Obj().Pkg().Name(), tvars.Zero)
					}
				}
			}

			err := tmpls.Lookup("workflow_two_return.tmpl").Execute(&buf, tvars)
			if err != nil {
				return nil, fmt.Errorf("error executing template for fn %s: %w", bfn.Fn.Name.String(), err)
			}
		default:
			return nil, fmt.Errorf("base workflow func %s must have either one or two return values", bfn.Fn.Name.String())
		}
	}

	genpath := filepath.Base(bf.Path)
	genpath = genpath[:len(genpath)-3] + ".workflow_gen.go"
	return codejen.NewFile(genpath, buf.Bytes(), WorkflowJenny{}), nil
}
