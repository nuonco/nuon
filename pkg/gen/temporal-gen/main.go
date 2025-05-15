package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"time"

	"github.com/grafana/codejen"
	"golang.org/x/tools/go/packages"
)

type BaseFile struct {
	// Path is the absolute path to the *ast.File
	Path string
	// File is the file containing funcs to be wrapped
	File *ast.File
	// ActivityFns is the list of ActivityFns in the file for which wrappers should be generated
	ActivityFns []ActivityFn
	// AsActivityFns is the list of ActivityFns in the file for which wrappers should be generated
	AsActivityFns []AsActivityFn
	// WorkflowFns is the list of WorkflowFns in the file for which wrappers should be generated
	WorkflowFns []WorkflowFn
	// Package is the result of loading and typechecking the package containing the funcs to be wrapped
	Package *packages.Package
}

// ActivityFn is the IR of a Go function to generate a wrapper for use as a Temporal activity. Generates only a wrapper
// for the calling (workflow) side.
type ActivityFn struct {
	// Fn is the node of the func for which a wrapper is to be generated
	Fn *ast.FuncDecl
	// Opts are options specified in comments on the func to be wrapped that modify generator output
	Opts *ActivityGenOptions
}

// ActivityGenOptions are specified as @-comments in the input Go source code, providing the user with
// control over various aspects of the generated output. This includes which temporal.ActivityOptions will
// be used in generated await functions
//
// TODO(sdboyer) expand this to mirror all the options Temporal actually exposes
type ActivityGenOptions struct {
	ScheduleToCloseTimeout time.Duration
	StartToCloseTimeout    time.Duration
	MaxRetries             int
	ByIdOnly               bool
	ById                   ByIdOptions
	OptionsCallback        string
}

type ByIdOptions struct {
	Name string
	Type string
}

// AsActivityFn is the IR of a Go function to be made usable as a Temporal activity by generating wrappers on both
// the workflow/calling side and the receiving invocation/activity side.
type AsActivityFn struct {
	// Fn is the node of the func for which a wrapper is to be generated
	Fn *ast.FuncDecl
	// Opts are options specified in comments on the func to be wrapped that modify generator output
	Opts *AsActivityGenOptions
}

type AsActivityGenOptions struct {
	Inner *ActivityGenOptions
}

// WorkflowFn is the IR of a Go function to be wrapped in generated code for use as a Temporal workflow
type WorkflowFn struct {
	// Fn is the node of the func for which a wrapper is to be generated
	Fn *ast.FuncDecl
	// Opts are options specified in comments on the func to be wrapped that modify generator output
	Opts *WorkflowGenOptions
}

// WorkflowGenOptions are specified as @-comments in the input Go source code, providing the user with
// control over various aspects of the generated output. This includes which temporal.ChildWorkflowOptions will
// be used in generated functions
type WorkflowGenOptions struct {
	ExecutionTimeout    time.Duration
	TaskTimeout         time.Duration
	IDCallback          string
	IDTemplate          string
	WaitForCancellation bool
	TaskQueue           string
	OptionsCallback     string
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
	bfs, err := parseDir(ctx, cwd)
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
		ActivityJenny{},       // O2O. Makes call-side await wrappers for activities
		WorkflowJenny{},       // O2O. Makes call-side await wrappers for workflows
		ActivityListJenny{},   // M2O. Makes Temporal registration-friendly list of all activities
		WorkflowListJenny{},   // M2O. Makes Temporal registration-friendly list of all workflows
		AsActivityJenny{},     // O2O. Makes impl-side wrapper for an activity
		AsActivityTypeJenny{}, // M2O. Makes struct, init fns to hold all impl-side wrappers
	)

	// Postprocessors are run on each output produced by the pipeline
	pipe.AddPostprocessors(GoImportsMapper, SlashHeaderMapper("mono/pkg/bins/temporal-gen"))

	// TODO(sdboyer) consolidate into one pipeline. This is currently two pipelines because of how the existing jennies are coupled to the BaseFile type:
	// the basic, OSS-able jennies take a BaseFile then pick the list they want to generate from. This works OK, but presents a problem for a standard
	// codejen.OneToOne adapter mode, because the adapter would have to mutate the BaseFile or create a whole new one with the unwrapped AsActivityFn list
	// for the basic jennies to consume.
	//
	// Better approach is probably to refactor the jennies to operate on either one or a list of the actual *Fn types they want, then wrap them with an
	// intermediate layer that decides on how to group them.
	//
	// Fortunately this is easy to hack around by simply having two pipelines that write to the same FS, and write the second pipeline such that
	// it just doesn't matter the inputs are being mutated.

	// New pipeline for As* jennies
	pipe2 := codejen.JennyListWithNamer(func(base *BaseFile) string {
		return filepath.Base(base.Path)
	})
	pipe2.Append(
		codejen.AdaptOneToOne(ActivityJenny{}, transformAsActivity),
	)
	pipe2.AddPostprocessors(GoImportsMapper, SlashHeaderMapper("mono/pkg/bins/temporal-gen"))
	// Run the pipelines with our inputs, and get a virtual FS back with all the outputs
	jfs, err := pipe.GenerateFS(bfs...)
	if err != nil {
		die(err)
	}

	jfs2, err := pipe2.GenerateFS(bfs...)
	if err != nil {
		die(err)
	}

	err = jfs.Merge(jfs2)
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
	return fmt.Errorf("%s:%d:%d: %w", pos.Filename, pos.Line, pos.Column, err)
}

func transformAsActivity(bf *BaseFile) *BaseFile {
	// TODO(sdboyer) find a more elegant way of hoisting this info than just reparsing output
	f, err := AsActivityJenny{}.Generate(bf)
	if err != nil {
		panic(err)
	}
	// TODO need to fix this in jenny system, returning a nil file ought to cause the underlying jenny to be skipped
	if f == nil {
		return nil
	}

	fset := token.NewFileSet()
	pf, err := parser.ParseFile(fset, f.RelativePath, f.Data, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	fns, gen := make(map[string]*ast.FuncDecl), make(map[string]*ast.StructType)
	for _, decl := range pf.Decls {
		switch x := decl.(type) {
		case *ast.FuncDecl:
			fns[x.Name.Name] = x
		case *ast.GenDecl:
			if x.Tok == token.TYPE &&
				len(x.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType).Fields.List) == 1 &&
				len(x.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType).Fields.List[0].Names) == 1 {
				// If it's a request struct type with only one field, then record the name of the type for later lookup
				gen[x.Specs[0].(*ast.TypeSpec).Name.Name] = x.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType)
			}
		}
	}

	bf.ActivityFns = make([]ActivityFn, 0, len(bf.AsActivityFns))
	for _, fn := range bf.AsActivityFns {
		if gfn, has := fns[fn.Fn.Name.Name]; has {
			cfg := ActivityFn{
				Fn:   gfn,
				Opts: fn.Opts.Inner,
			}

			// Look up against the request struct type map
			if greq, has := gen[gfn.Type.Params.List[1].Type.(*ast.Ident).Name]; has {
				// if we have a match, then we can make a shortened by id fn
				cfg.Opts.ById = ByIdOptions{
					Name: greq.Fields.List[0].Names[0].Name,
					Type: greq.Fields.List[0].Type.(*ast.Ident).Name,
				}
				// And we don't want the other function exposed at all
				cfg.Opts.ByIdOnly = true
			}

			bf.ActivityFns = append(bf.ActivityFns, cfg)
		} else {
			panic(fmt.Errorf("could not find generated function %q", fn.Fn.Name.Name))
		}
		// TODO transform the Fn's args into the request type we'll be using
	}

	return bf
}
