package file

import (
	"fmt"
	"go/ast"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/dir"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/labels"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/parser"
)

// File represents a single file that may contain generated code
type File struct {
	Path        string
	Package     *dir.Package
	Annotations []*parser.Annotation
	Functions   []*Function
}

// Function represents a function declaration that was annotated
type Function struct {
	Decl       *ast.FuncDecl
	Annotation *parser.Annotation
}

// ProcessFile scans a file for annotations and returns a File struct if any are found.
// cfg supplies label defaults and may be nil when no temporal-gen.yaml is in use.
func ProcessFile(pkg *dir.Package, file *ast.File, path string, strict bool, cfg *labels.Config) (*File, error) {
	var functions []*Function
	var parseErr error

	ast.Inspect(file, func(n ast.Node) bool {
		if parseErr != nil {
			return false
		}

		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Doc == nil {
			return true
		}

		// Extract comments
		var comments []string
		for _, c := range fn.Doc.List {
			comments = append(comments, c.Text)
		}

		// Parse annotations
		annotation, err := parser.ParseWithLabels(comments, cfg)
		if err != nil {
			if strict {
				parseErr = fmt.Errorf("error parsing annotations in function %s: %w", fn.Name.Name, err)
				return false
			}
			fmt.Printf("Warning: error parsing annotations in function %s: %v\n", fn.Name.Name, err)
			return true
		}

		if annotation != nil {
			functions = append(functions, &Function{
				Decl:       fn,
				Annotation: annotation,
			})
		}

		return true
	})

	if parseErr != nil {
		return nil, parseErr
	}

	if len(functions) == 0 {
		return nil, nil
	}

	return &File{
		Path:      path,
		Package:   pkg,
		Functions: functions,
	}, nil
}
