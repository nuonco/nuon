package generator

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessFile(t *testing.T) {
	src := `package test
	
	// @` + config.AnnotationPrefix + ` activity
	// @schedule-to-close-timeout 1h
	func MyActivity(ctx context.Context, input string) error {
		return nil
	}

	// This is not annotated
	func OtherFunc() {}
	`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	pkg := &Package{} // Mock package for now

	result, err := ProcessFile(pkg, f, "test.go", true)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Len(t, result.Functions, 1)
	assert.Equal(t, "MyActivity", result.Functions[0].Decl.Name.Name)
	assert.Equal(t, "activity", result.Functions[0].Annotation.Type)
}
