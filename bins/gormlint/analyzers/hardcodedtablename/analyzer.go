package hardcodedtablename

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "hardcodedtablename",
	Doc:      "reports raw SQL strings in GORM Joins, Order, Select, Group, and Having calls",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

var flaggedMethods = map[string]string{
	"Joins":  "use views.TableOrViewName() or association joins instead of raw SQL join strings",
	"Order":  "use a constant or views.TableOrViewName() for column references",
	"Select": "use a constant or explicit column list for column references",
	"Group":  "use a constant or views.TableOrViewName() for column references",
	"Having": "use a constant or views.TableOrViewName() for column references",
}

func run(pass *analysis.Pass) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		suggestion, ok := flaggedMethods[sel.Sel.Name]
		if !ok {
			return
		}

		if len(call.Args) == 0 {
			return
		}

		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return
		}

		val := strings.Trim(lit.Value, "`\"")

		if sel.Sel.Name == "Joins" && !looksLikeSQL(val) {
			return
		}

		pass.Reportf(lit.Pos(), "raw SQL string in %s(); %s", sel.Sel.Name, suggestion)
	})

	return nil, nil
}

func looksLikeSQL(s string) bool {
	upper := strings.ToUpper(s)
	for _, kw := range []string{"JOIN", " ON ", "SELECT", "FROM", "WHERE"} {
		if strings.Contains(upper, kw) {
			return true
		}
	}
	return false
}
