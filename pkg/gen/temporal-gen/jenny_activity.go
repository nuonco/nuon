package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"
	"time"

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

	// ReqIsPtr indicates whether the type of the request value is a pointer
	ReqIsPtr bool

	// Options contains the input-specified options that set specify ActivityOptions and govern
	// various other generator behaviors
	Options ActivityGenOptions

	// Zero is a string literal representing the zero value of the response type
	Zero string
}

type tvars_activityfn_byid struct {
	ByIdFnName  string
	AwaitFnName string
	RespType    string
	ReqType     string
	ReqIsPtr    bool
	IdFieldName string
	IdType      string
	ByIdOnly    bool
	BaseFnName  string
}

// ActivityJenny is a jenny that generates a function that calls a provided base function as a Temporal activity, then awaits the response.
type ActivityJenny struct {
	UseMethods bool
}

func (w ActivityJenny) JennyName() string {
	return "ActivityJenny"
}

func (w ActivityJenny) Generate(bf *BaseFile) (*codejen.File, error) {
	if bf == nil || len(bf.ActivityFns) == 0 {
		return nil, nil
	}

	var buf bytes.Buffer
	genImports(&buf, bf)
	for _, bfn := range bf.ActivityFns {
		var wv tvars_awaitfn

		wv.Options = *bfn.Opts

		if wv.Options.StartToCloseTimeout == 0 {
			wv.Options.StartToCloseTimeout = 5 * time.Second
		}
		if wv.Options.ScheduleToCloseTimeout == 0 {
			wv.Options.ScheduleToCloseTimeout = 30 * time.Minute
			if wv.Options.StartToCloseTimeout > wv.Options.ScheduleToCloseTimeout {
				wv.Options.ScheduleToCloseTimeout = wv.Options.StartToCloseTimeout
			}
		}

		bfname := bfn.Fn.Name.String()
		lead := "A"
		if bfn.Opts.ById.Name != "" && bfn.Opts.ByIdOnly {
			lead = "a"
		}
		wv.FnName = fmt.Sprintf("%swait%s", lead, strings.Title(bfname))

		wv.ReqType = astfmt.Sprint(bfn.Fn.Type.Params.List[1].Type)
		_, wv.ReqIsPtr = bfn.Fn.Type.Params.List[1].Type.(*ast.StarExpr)

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
			wv.IsMethod = w.UseMethods
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
			err := tmpls.Lookup("activity_one_return.tmpl").Execute(&buf, wv)
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
				// Non-pointer return types need to be zero-initialized, the syntax for which
				// varies by type
				rt := bf.Package.TypesInfo.Types[respt].Type
				wv.Zero = zerostr(rt)

				// If the return type is a struct imported from a different package, we have to qualify it
				if x, ok := rt.(*types.Named); ok && x.Obj().Pkg().Path() != bf.Package.PkgPath {
					if _, is := x.Underlying().(*types.Struct); is {
						wv.Zero = fmt.Sprintf("%s.%s", x.Obj().Pkg().Name(), wv.Zero)
					}
				}
			}
			err := tmpls.Lookup("activity_two_return.tmpl").Execute(&buf, wv)
			if err != nil {
				return nil, fmt.Errorf("error executing template for fn %s: %w", bfn.Fn.Name.String(), err)
			}
		default:
			return nil, fmt.Errorf("base activity func %s must have either one or two return values", bfn.Fn.Name.String())
		}

		if bfn.Opts.ById.Name != "" {
			idopts := tvars_activityfn_byid{
				ByIdFnName:  fmt.Sprintf("%sBy%s", wv.FnName, bfn.Opts.ById.Name),
				AwaitFnName: wv.FnName,
				RespType:    wv.RespType,
				ReqType:     wv.ReqType,
				ReqIsPtr:    wv.ReqIsPtr,
				IdFieldName: bfn.Opts.ById.Name,
				IdType:      bfn.Opts.ById.Type,
				BaseFnName:  wv.BaseFnName,
				ByIdOnly:    bfn.Opts.ByIdOnly,
			}

			if bfn.Opts.ByIdOnly {
				idopts.ByIdFnName = "A" + wv.FnName[1:]
			}
			err := tmpls.Lookup("activity_by_id.tmpl").Execute(&buf, idopts)
			if err != nil {
				return nil, fmt.Errorf("error executing by_id template for fn %s: %w", bfn.Fn.Name.String(), err)
			}
		}
	}

	genpath := filepath.Base(bf.Path)
	genpath = genpath[:len(genpath)-3] + ".activity_gen.go"
	return codejen.NewFile(genpath, buf.Bytes(), ActivityJenny{}), nil
}
