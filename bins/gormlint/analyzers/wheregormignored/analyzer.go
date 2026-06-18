package wheregormignored

import (
	"go/ast"
	"go/types"
	"reflect"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "wheregormignored",
	Doc:      "reports struct fields tagged gorm:\"-\" used in GORM Where clauses",
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

		lit, ok := call.Args[0].(*ast.CompositeLit)
		if !ok {
			// Also handle &Struct{} (address-of composite literal)
			unary, ok := call.Args[0].(*ast.UnaryExpr)
			if !ok {
				return
			}
			lit, ok = unary.X.(*ast.CompositeLit)
			if !ok {
				return
			}
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
			if tag == "" {
				continue
			}

			gormTag := reflect.StructTag(tag).Get("gorm")
			if gormTag == "-" {
				pass.Reportf(kv.Pos(), "field %q has gorm:\"-\" tag and will be silently ignored in this Where clause", ident.Name)
			}
		}
	})

	return nil, nil
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
