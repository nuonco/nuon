package missingcontext

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "missingcontext",
	Doc:      "reports GORM query chains that lack WithContext(), losing request tracing and cancellation",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

var terminalMethods = map[string]bool{
	"Find":    true,
	"First":   true,
	"Last":    true,
	"Take":    true,
	"Create":  true,
	"Save":    true,
	"Update":  true,
	"Updates": true,
	"Delete":  true,
	"Scan":    true,
	"Count":   true,
	"Exec":    true,
	"Raw":     true,
	"Pluck":   true,
}

var gormChainMethods = map[string]bool{
	"Where":       true,
	"Preload":     true,
	"Model":       true,
	"Joins":       true,
	"Order":       true,
	"Select":      true,
	"Group":       true,
	"Having":      true,
	"Limit":       true,
	"Offset":      true,
	"Scopes":      true,
	"Or":          true,
	"Not":         true,
	"WithContext": true,
}

var hookReceiverParams = map[string]bool{
	"BeforeCreate": true,
	"AfterCreate":  true,
	"BeforeUpdate": true,
	"AfterUpdate":  true,
	"BeforeDelete": true,
	"AfterDelete":  true,
	"BeforeSave":   true,
	"AfterSave":    true,
	"AfterQuery":   true,
	"AfterFind":    true,
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

		if chainContainsWithContext(sel.X) {
			return
		}

		root := chainRoot(sel.X)
		if root == nil {
			return
		}

		if !isGormDBField(pass, root) {
			return
		}

		if isInsideGormHook(pass, call) {
			return
		}

		pass.Reportf(call.Pos(), "GORM query chain missing WithContext(); add .WithContext(ctx) for request tracing and cancellation support")
	})

	return nil, nil
}

func chainContainsWithContext(expr ast.Expr) bool {
	for {
		call, ok := expr.(*ast.CallExpr)
		if !ok {
			return false
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		if sel.Sel.Name == "WithContext" {
			return true
		}
		if !gormChainMethods[sel.Sel.Name] && !terminalMethods[sel.Sel.Name] {
			return false
		}
		expr = sel.X
	}
}

func chainRoot(expr ast.Expr) ast.Expr {
	for {
		call, ok := expr.(*ast.CallExpr)
		if !ok {
			return expr
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return expr
		}
		if !gormChainMethods[sel.Sel.Name] && !terminalMethods[sel.Sel.Name] {
			return expr
		}
		expr = sel.X
	}
}

func isGormDBField(pass *analysis.Pass, expr ast.Expr) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if sel.Sel.Name != "db" && sel.Sel.Name != "DB" {
		return false
	}

	typ := pass.TypesInfo.TypeOf(expr)
	if typ == nil {
		return false
	}

	return isGormDBType(typ)
}

func isGormDBType(t types.Type) bool {
	ptr, ok := t.(*types.Pointer)
	if ok {
		t = ptr.Elem()
	}

	named, ok := t.(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	return obj.Name() == "DB" && obj.Pkg() != nil && obj.Pkg().Path() == "gorm.io/gorm"
}

func isInsideGormHook(pass *analysis.Pass, node ast.Node) bool {
	for _, f := range pass.Files {
		for _, decl := range f.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Recv == nil {
				continue
			}

			if hookReceiverParams[funcDecl.Name.Name] && nodeWithin(node, funcDecl) {
				return true
			}
		}
	}
	return false
}

func nodeWithin(node ast.Node, funcDecl *ast.FuncDecl) bool {
	return node.Pos() >= funcDecl.Body.Pos() && node.End() <= funcDecl.Body.End()
}
