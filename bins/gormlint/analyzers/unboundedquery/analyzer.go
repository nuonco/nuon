package unboundedquery

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "unboundedquery",
	Doc:      "reports GORM Find/Scan calls without a Limit in the query chain",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

var terminalMethods = map[string]bool{
	"Find": true,
	"Scan": true,
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		if !terminalMethods[sel.Sel.Name] {
			return
		}

		if sel.Sel.Name == "Find" && len(call.Args) > 1 {
			return
		}

		if chainContains(sel.X, "Limit") {
			return
		}

		if chainContains(sel.X, "First") || chainContains(sel.X, "Last") || chainContains(sel.X, "Take") {
			return
		}

		pass.Reportf(call.Pos(), "%s() without Limit() in query chain; consider adding Limit() to prevent loading all matching records", sel.Sel.Name)
	})

	return nil, nil
}

func chainContains(expr ast.Expr, methodName string) bool {
	for {
		call, ok := expr.(*ast.CallExpr)
		if !ok {
			return false
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		if sel.Sel.Name == methodName {
			return true
		}
		expr = sel.X
	}
}
