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

// tvars_awaitfn contains all the template variables used to generate an await function.
type tvars_awaitfn struct {
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

	// Options contains the input-specified options that set specify ActivityOptions
	Options GenOptions

	// Zero is a string literal representing the zero value of the response type
	Zero string
}

// AwaitJenny is a jenny that generates a function that calls another function as a Temporal activity, then awaits the response.
type AwaitJenny struct{}

func (w AwaitJenny) JennyName() string {
	return "AwaitJenny"
}

func (w AwaitJenny) Generate(bf *BaseFile) (*codejen.File, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, `package %s

import (
`, bf.File.Name.Name)

	imports := map[string]bool{
		"time":                        true,
		"go.temporal.io/sdk/temporal": true,
		"go.temporal.io/sdk/workflow": true,
	}

	// pull in all the imports from the file containing the base func, so that at minimum we have the right import for
	// the request and response types. unnecessary imports will be removed by a goimports pass later. We do deduplicate,
	// though, because it appears that goimports has a bug and doesn't reliably dedup
	for _, im := range bf.File.Imports {
		imports[strings.Trim(astfmt.Sprint(im), "\"")] = true
	}
	for im := range imports {
		fmt.Fprintf(&buf, "\n\t%q", im)
	}
	buf.WriteString("\n)")

	for _, bfn := range bf.Fns {
		var wv tvars_awaitfn

		wv.Options = *bfn.Opts
		bfname := bfn.Fn.Name.String()
		wv.FnName = fmt.Sprintf("Await%s", strings.Title(bfname))

		wv.ReqType = astfmt.Sprint(bfn.Fn.Type.Params.List[1].Type)

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
			wv.IsMethod = true
			recvVar := bfn.Fn.Recv.List[0].Names[0].Name
			wv.Receiver = fmt.Sprintf("%s %s", recvVar, astfmt.Sprint(bfn.Fn.Recv.List[0].Type))

			wv.BaseFnName = fmt.Sprintf("%s.%s", astfmt.Sprint(bfn.Fn.Recv.List[0].Type), bfname)
		} else {
			wv.BaseFnName = bfname
		}
		basebuf.WriteString(bfname)

		wv.BaseFnSymbol = basebuf.String()
		switch len(bfn.Fn.Type.Results.List) {
		case 1:
			// This template assumes without verification that the wrapped fn is has a bare error return, b/c temporal requires that
			err := tmpls.Lookup("one_return.tmpl").Execute(&buf, wv)
			if err != nil {
				return nil, fmt.Errorf("error executing template for fn %s: %w", bfn.Fn.Name.String(), err)
			}
		case 2:
			// This template assumes without verification that the wrapped fn is a (custom value, error) return, b/c temporal requires that
			wv.HasResp = true
			respt := bfn.Fn.Type.Results.List[0].Type
			wv.RespType = astfmt.Sprint(respt)
			_, wv.RespIsPtr = bfn.Fn.Type.Results.List[0].Type.(*ast.StarExpr)

			if !wv.RespIsPtr {
				rt := bf.Package.TypesInfo.Types[respt].Type
				if rtn, ok := rt.(*types.Named); ok {
					switch rtn.Underlying().(type) {
					case *types.Struct:
						wv.Zero = fmt.Sprintf("%s{}", rtn.Obj().Name())
					default:
						wv.Zero = zerostr(rtn.Underlying())
					}
				} else {
					wv.Zero = zerostr(rt)
				}
			}
			err := tmpls.Lookup("two_return.tmpl").Execute(&buf, wv)
			if err != nil {
				return nil, fmt.Errorf("error executing template for fn %s: %w", bfn.Fn.Name.String(), err)
			}
		default:
			return nil, fmt.Errorf("base activity func %s must have either one or two return values", bfn.Fn.Name.String())
		}
	}

	genpath := filepath.Base(bf.Path)
	genpath = genpath[:len(genpath)-3] + "_gen.go"
	return codejen.NewFile(genpath, buf.Bytes(), AwaitJenny{}), nil
}

func zerostr(rt types.Type) string {
	// TODO(sdboyer) move checking over to the parser to error earlier if we have these return types
	switch x := rt.(type) {
	case *types.Named:
		panic("unwrap from the named type before calling this")
	case *types.Basic:
		if x.Info()&types.IsNumeric != 0 {
			return "0"
		}
		if x.Info()&types.IsString != 0 {
			return "\"\""
		}
		if x.Info()&types.IsBoolean != 0 {
			return "false"
		}
		panic(fmt.Sprintf("unhandled zero value generation for basic type: %s", x.String()))
	case *types.Array, *types.Slice, *types.Map, *types.Chan:
		return "nil"
	case *types.Struct:
		return "struct{}{}"
	}
	if x, ok := rt.(*types.Named); ok {
		return x.Obj().Name()
	}
	panic(fmt.Errorf("unhandled zero value generation for type: %s", rt.String()))
}
