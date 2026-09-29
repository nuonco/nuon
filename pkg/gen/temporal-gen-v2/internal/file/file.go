package file

import (
	"errors"
	"fmt"
	"go/ast"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/dir"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/parser"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/tags"
)

type File struct {
	Path        string
	Package     *dir.Package
	Annotations []*parser.Annotation
	Functions   []*Function
}

type Function struct {
	Decl       *ast.FuncDecl
	Annotation *parser.Annotation
}

func ProcessFile(pkg *dir.Package, file *ast.File, path string, strict bool, cfg *tags.Config) (*File, error) {
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

		var comments []string
		for _, c := range fn.Doc.List {
			comments = append(comments, c.Text)
		}

		annotation, err := parser.ParseWithTags(comments, cfg)
		if err != nil {
			// why: A tag that cannot be resolved is fatal even when not strict:
			// downgrading it to a warning would skip the function and silently
			// drop a wrapper the caller expects to exist.
			var tagErr *parser.TagError
			if strict || errors.As(err, &tagErr) {
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
