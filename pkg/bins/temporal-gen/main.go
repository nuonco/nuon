package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"time"

	"github.com/grafana/codejen"
)

type BaseFile struct {
	// Path is the absolute path to the *ast.File
	Path string
	// File is the file containing funcs to be wrapped
	File *ast.File
	// Fns is the list of BaseFns in the file for which wrappers should be generated
	Fns []BaseFn
}

// BaseFn is the IR of a Go function to be wrapped in generated code for use as a Temporal activity
type BaseFn struct {
	// Fn is the node of the func for which a wrapper is to be generated
	Fn *ast.FuncDecl
	// Opts are options specified in comments on the func to be wrapped that modify generator output
	Opts *GenOptions
}

// GenOptions are specified as @-comments in the input Go source code, and determine which temporal.ActivityOptions will
// be used in generated await functions.
//
// TODO(sdboyer) expand this to mirror all the options Temporal actually exposes
type GenOptions struct {
	Timeout    time.Duration
	MaxRetries int
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get working directory: %s", err)
		os.Exit(1)
	}
	if len(os.Args) > 1 {
		fmt.Fprintf(os.Stderr, "code generator does not currently accept any arguments\n, got %q", os.Args)
		os.Exit(1)
	}

	ctx := context.Background()

	// Parse files in the cwd to find inputs to our generator
	bfs, err := loadBase(ctx, cwd)
	if err != nil {
		die(err)
	}

	// Create the base jenny pipeline
	pipe := codejen.JennyListWithNamer(func(base *BaseFile) string {
		return filepath.Base(base.Path)
	})

	// Add our await jenny. Its signature is a codejen.OneToOne, which means it will be called once
	// for each item in our inputs, and will produce one file for each of those inputs
	pipe.Append(
		AwaitJenny{},
	)

	// Postprocessors are run on each output produced by the pipeline
	pipe.AddPostprocessors(GoImportsMapper, SlashHeaderMapper("mono/bins/temporal-gen"))

	// Run the pipeline with our inputs, and get a virtual FS back with all the outputs
	jfs, err := pipe.GenerateFS(bfs...)
	if err != nil {
		die(err)
	}

	// Nuon doesn't commit generated files, but if there's ever a need to compare/check files on disk against the
	// generated files, this can be uncommented into an if/else with jfs.Write
	//
	// if _, set := os.LookupEnv("CODEGEN_VERIFY"); set {
	// 	if err = jfs.Verify(ctx, cwd); err != nil {
	// 		die(fmt.Errorf("generated code is out of sync with inputs:\n%s\nrun `make gen-cue` to regenerate", err))
	// 	}
	// }

	if err = jfs.Write(ctx, cwd); err != nil {
		die(fmt.Errorf("error while writing generated code to disk:\n%s", err))
	}
}

func die(err error) {
	fmt.Fprint(os.Stderr, err, "\n")
	os.Exit(1)
}

func withPos(fset *token.FileSet, tok token.Pos, err error) error {
	pos := fset.Position(tok)
	return fmt.Errorf("%s:%d:%d: %w\n", pos.Filename, pos.Line, pos.Column, err)
}
