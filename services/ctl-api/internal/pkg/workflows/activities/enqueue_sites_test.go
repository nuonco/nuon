package activities_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
)

func TestQueueEnqueueSitesAreExplicit(t *testing.T) {
	root := filepath.Clean("../../..")
	constants := stringConstants(t, root)

	var violations []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch typed := node.(type) {
			case *ast.CompositeLit:
				if typeName(typed.Type) != "EnqueueSignalToOwnerRequest" {
					return true
				}
				if !hasField(typed, "QueueName") && !hasField(typed, "QueueID") {
					violations = append(violations, fset.Position(typed.Pos()).String()+": enqueue has no queue name or ID")
				}
				if _, literal := fieldValue(typed, "QueueName").(*ast.BasicLit); literal {
					violations = append(violations, fset.Position(typed.Pos()).String()+": queue name must use a queuenames constant")
				}
				ownerType, hasOwnerType := resolveString(fieldValue(typed, "OwnerType"), constants)
				queueName, hasQueueName := resolveString(fieldValue(typed, "QueueName"), constants)
				if hasOwnerType && hasQueueName {
					if _, ok := queuenames.SpecByName(ownerType, queueName); !ok {
						violations = append(violations, fset.Position(typed.Pos()).String()+": queue name is not registered for owner type")
					}
				}
			case *ast.CallExpr:
				selector, ok := typed.Fun.(*ast.SelectorExpr)
				if ok && (selector.Sel.Name == "GetQueueByOwner" || selector.Sel.Name == "AwaitGetQueueByOwner") {
					violations = append(violations, fset.Position(typed.Pos()).String()+": ambiguous queue lookup by owner")
				}
			}
			return true
		})
		return nil
	})
	require.NoError(t, err)
	require.Empty(t, violations)
}

// stringConstants maps every string constant declared under root to its value,
// keyed by identifier. Queue names reach enqueue sites through package aliases
// (appshelpers.AppInstallSyncsQueueName and friends), so resolving by name lets
// the guard follow them without type-checking the whole tree.
func stringConstants(t *testing.T, root string) map[string]string {
	t.Helper()

	exprs := map[string]ast.Expr{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.CONST {
				continue
			}
			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for idx, name := range valueSpec.Names {
					if idx < len(valueSpec.Values) {
						exprs[name.Name] = valueSpec.Values[idx]
					}
				}
			}
		}
		return nil
	})
	require.NoError(t, err)

	values := map[string]string{}
	for {
		progressed := false
		for name, expr := range exprs {
			if _, done := values[name]; done {
				continue
			}
			value, ok := resolveString(expr, values)
			if !ok {
				continue
			}
			values[name] = value
			progressed = true
		}
		if !progressed {
			return values
		}
	}
}

func resolveString(expr ast.Expr, constants map[string]string) (string, bool) {
	switch typed := expr.(type) {
	case *ast.BasicLit:
		if typed.Kind != token.STRING {
			return "", false
		}
		unquoted, err := strconv.Unquote(typed.Value)
		if err != nil {
			return "", false
		}
		return unquoted, true
	case *ast.Ident:
		value, ok := constants[typed.Name]
		return value, ok
	case *ast.SelectorExpr:
		value, ok := constants[typed.Sel.Name]
		return value, ok
	default:
		return "", false
	}
}

func fieldValue(lit *ast.CompositeLit, name string) ast.Expr {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		ident, ok := kv.Key.(*ast.Ident)
		if ok && ident.Name == name {
			return kv.Value
		}
	}
	return nil
}

func typeName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		return typed.Sel.Name
	case *ast.StarExpr:
		return typeName(typed.X)
	default:
		return ""
	}
}

func hasField(lit *ast.CompositeLit, name string) bool {
	return fieldValue(lit, name) != nil
}
