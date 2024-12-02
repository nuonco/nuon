package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"strings"

	"github.com/go-toolsmith/astfmt"
)

type methodSymbols struct {
	receiverLit      string
	receiverVar      string
	receiverType     string
	receiverTypeName string
	fnSym            string
}

func extractFnSymbols(fn *ast.FuncDecl) methodSymbols {
	var ret methodSymbols
	bfname := fn.Name.String()

	basebuf := new(strings.Builder)
	if fn.Recv != nil {
		ret.receiverType = astfmt.Sprint(fn.Recv.List[0].Type)
		if x, is := fn.Recv.List[0].Type.(*ast.StarExpr); is {
			ret.receiverTypeName = astfmt.Sprint(x.X)
		} else {
			ret.receiverTypeName = ret.receiverType
		}
		ret.receiverLit = fmt.Sprintf("%s %s", fn.Recv.List[0].Names[0].Name, ret.receiverType)
		ret.receiverVar = fn.Recv.List[0].Names[0].Name

		basebuf.WriteString("(")
		expr := fn.Recv.List[0].Type
		if x, isPtr := expr.(*ast.StarExpr); isPtr {
			expr = x.X
			basebuf.WriteString("&")
		}
		// Injecting {} hardcodes assumption that receiver is struct-kinded
		basebuf.WriteString(astfmt.Sprint(expr) + "{}).")

		// wv.BaseFnName = fmt.Sprintf("%s.%s", astfmt.Sprint(fn.Recv.List[0].Type), bfname)
	}
	basebuf.WriteString(bfname)

	ret.fnSym = basebuf.String()
	return ret
}

func genImports(buf *bytes.Buffer, bf *BaseFile) {
	fmt.Fprintf(buf, `package %s

import (
`, bf.File.Name.Name)

	q := func(s string) string { return "\"" + s + "\"" }
	imports := map[string]bool{
		q("time"):                                        true,
		q("go.temporal.io/sdk/temporal"):                 true,
		q("go.temporal.io/sdk/workflow"):                 true,
		q("github.com/powertoolsdev/mono/pkg/workflows"): true,
		q("github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"): true,
		"enumsv1 " + q("go.temporal.io/api/enums/v1"):                              true,
	}

	// pull in all the imports from the file containing the base func, so that at minimum we have the right import for
	// the request and response types. unnecessary imports will be removed by a goimports pass later. We do deduplicate,
	// though, because it appears that goimports has a bug and doesn't reliably dedup
	for _, im := range bf.File.Imports {
		// NOTE(sdboyer) this will generate invalid stuff if there's a named import that is in the static list above
		imports[astfmt.Sprint(im)] = true
	}
	for im := range imports {
		fmt.Fprintf(buf, "\n\t%s", im)
	}
	buf.WriteString("\n)")
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
