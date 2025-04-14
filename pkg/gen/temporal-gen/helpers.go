package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
	"unicode"

	"github.com/go-toolsmith/astfmt"
)

type methodSymbols struct {
	receiverLit      string
	receiverVar      string
	receiverType     string
	receiverTypeName string
	fnSym            string
}

func activityTypeFor(fn *ast.FuncDecl) string {
	if fn.Recv != nil {
		var recvname string
		if x, is := fn.Recv.List[0].Type.(*ast.StarExpr); is {
			recvname = astfmt.Sprint(x.X)
		} else {
			recvname = astfmt.Sprint(fn.Recv.List[0].Type)
		}
		return fmt.Sprintf("%sActivities", recvname)
	} else {
		return "Activities"
	}
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
		switch x.Underlying().(type) {
		case *types.Struct:
			return fmt.Sprintf("%s{}", x.Obj().Name())
		default:
			return zerostr(x.Underlying())
		}
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
	panic(fmt.Errorf("unhandled zero value generation for type: %s", rt.String()))
}

func paramsToStruct(fset *token.FileSet, fn *ast.FuncDecl) (*ast.GenDecl, error) {
	if len(fn.Type.Params.List) < 2 {
		return nil, withPos(fset, fn.Type.Params.Pos(), fmt.Errorf("functions annotated with as-activity must have at least two parameters"))
	}

	ret := &ast.StructType{
		Fields: &ast.FieldList{},
	}
	gd := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{
					Name: fmt.Sprintf("%sRequest", fn.Name.Name),
				},
				Type: ret,
			},
		},
	}

	for _, param := range fn.Type.Params.List[1:] {
		var ns []*ast.Ident
		for _, name := range param.Names {
			ns = append(ns, titleize(name))
		}

		ret.Fields.List = append(ret.Fields.List, &ast.Field{
			Names: ns,
			Type:  param.Type,
		})

		// if len(param.Names) > 1 {
		// 	fmt.Println(param.Names)
		// 	return nil, withPos(fset, param.Pos(), fmt.Errorf("as-activity params must be named"))
		// }
	}

	return gd, nil
}

func titleize(ident *ast.Ident) *ast.Ident {
	if ident == nil {
		return nil
	}

	// Split by underscores
	words := strings.Split(ident.Name, "_")
	for i, word := range words {
		if len(word) > 0 {
			// Title case each word
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return &ast.Ident{Name: strings.Join(words, "")}
}
