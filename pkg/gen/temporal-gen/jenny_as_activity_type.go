package main

import (
	"bytes"
	"fmt"
	"go/ast"

	"github.com/grafana/codejen"
)

type tvars_asactivity_type struct {
	// ActivitiesTypeName is the name of the generated struct.
	ActivitiesTypeName string

	// BaseTypeName is the name of the type whose methods are being wrapped on this activities type.
	BaseTypeName string
}

type AsActivityTypeJenny struct {}

func (j AsActivityTypeJenny) JennyName() string {
	return "AsActivityTypeJenny"
}

func (w AsActivityTypeJenny) Generate(bfs ...*BaseFile) (*codejen.File, error) {
	// TODO produce multiple if multiple groups are encountered
	var fn *ast.FuncDecl
	for _, bf := range bfs {
		if len(bf.AsActivityFns) > 0 {
			fn = bf.AsActivityFns[0].Fn
			break
		}
	}

	if fn == nil {
		return nil, nil
	}

	var buf bytes.Buffer
	// TODO merge imports
	genImports(&buf, bfs[0])

	var wv tvars_asactivity_type
	wv.ActivitiesTypeName = activityTypeFor(fn)
	wv.BaseTypeName = fn.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name

	err := tmpls.Lookup("as_activity_type.tmpl").Execute(&buf, wv)
	if err != nil {
		return nil, fmt.Errorf("error executing as_activity_type template for base type %s: %w", wv.BaseTypeName, err)
	}

	return codejen.NewFile("activities_type_gen.go", buf.Bytes(), AsActivityJenny{}), nil
}
