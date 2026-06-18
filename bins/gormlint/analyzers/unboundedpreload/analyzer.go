package unboundedpreload

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "unboundedpreload",
	Doc:      "reports GORM Preload calls without a scoping function containing Limit",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Preload" {
			return
		}

		if len(call.Args) == 0 {
			return
		}

		hasScopingFunc := false
		hasLimit := false

		for _, arg := range call.Args[1:] {
			funcLit, ok := arg.(*ast.FuncLit)
			if !ok {
				continue
			}
			hasScopingFunc = true
			hasLimit = containsLimitCall(funcLit.Body)
		}

		if !hasScopingFunc {
			pass.Reportf(call.Pos(), "unbounded Preload without scoping function; add a func(db *gorm.DB) *gorm.DB with Limit() to prevent loading all related records")
			return
		}

		if !hasLimit {
			pass.Reportf(call.Pos(), "Preload scoping function has no Limit(); consider adding Limit() to prevent loading all related records")
		}
	})

	return nil, nil
}

func containsLimitCall(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if sel.Sel.Name == "Limit" {
			found = true
			return false
		}
		return true
	})
	return found
}
