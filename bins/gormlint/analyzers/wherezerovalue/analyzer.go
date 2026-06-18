package wherezerovalue

import (
	"go/ast"
	"go/token"
	"go/types"
	"reflect"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "wherezerovalue",
	Doc:      "reports literal zero-value fields in GORM Where struct clauses that will be silently ignored",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Where" {
			return
		}

		if len(call.Args) == 0 {
			return
		}

		lit := extractCompositeLit(call.Args[0])
		if lit == nil {
			return
		}

		typ := pass.TypesInfo.TypeOf(lit)
		if typ == nil {
			return
		}

		structType := underlyingStruct(typ)
		if structType == nil {
			return
		}

		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			ident, ok := kv.Key.(*ast.Ident)
			if !ok {
				continue
			}

			fieldIdx := fieldIndex(structType, ident.Name)
			if fieldIdx < 0 {
				continue
			}

			tag := structType.Tag(fieldIdx)
			if tag != "" {
				gormTag := reflect.StructTag(tag).Get("gorm")
				if gormTag == "-" {
					continue
				}
			}

			if isLiteralZero(kv.Value) {
				pass.Reportf(kv.Pos(), "field %q has zero value and will be silently ignored by GORM in this Where clause; use a pointer type or map-based Where", ident.Name)
			}
		}
	})

	return nil, nil
}

func isLiteralZero(expr ast.Expr) bool {
	switch v := expr.(type) {
	case *ast.BasicLit:
		switch v.Kind {
		case token.INT:
			return v.Value == "0"
		case token.FLOAT:
			return v.Value == "0.0" || v.Value == "0."
		case token.STRING:
			return v.Value == `""` || v.Value == "``"
		}
	case *ast.Ident:
		return v.Name == "false"
	}
	return false
}

func extractCompositeLit(expr ast.Expr) *ast.CompositeLit {
	if lit, ok := expr.(*ast.CompositeLit); ok {
		return lit
	}
	if unary, ok := expr.(*ast.UnaryExpr); ok {
		if lit, ok := unary.X.(*ast.CompositeLit); ok {
			return lit
		}
	}
	return nil
}

func underlyingStruct(t types.Type) *types.Struct {
	t = t.Underlying()
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem().Underlying()
	}
	s, _ := t.(*types.Struct)
	return s
}

func fieldIndex(s *types.Struct, name string) int {
	for i := 0; i < s.NumFields(); i++ {
		if s.Field(i).Name() == name {
			return i
		}
	}
	return -1
}
