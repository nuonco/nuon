// Package temporalgen provides a library interface for temporal-gen-v2.
// It can be used programmatically rather than only via the CLI.
package temporalgen

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/dir"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/file"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/generator"
)

// Options configures a code generation run.
type Options struct {
	// Dir is the root directory to process. Defaults to ".".
	Dir string

	// Recursive processes all packages under Dir recursively (equivalent to ./...).
	Recursive bool

	// Validate fails if any annotation validation errors are found.
	Validate bool

	// Imports runs golang.org/x/tools/imports on generated files.
	Imports bool

	// Parallelism controls how many packages are processed concurrently within
	// each dependency level. Defaults to runtime.NumCPU() when <= 0.
	Parallelism int

	// OnPackage is called before processing each package. May be nil.
	OnPackage func(pkgName string)
}

// Generate runs temporal code generation with the provided options.
// It loads all packages in a single packages.Load call, resolves dependency
// ordering, and processes packages in parallel within each level.
func Generate(ctx context.Context, opts Options) error {
	targetDir := opts.Dir
	if targetDir == "" {
		targetDir = "."
	}

	parallelism := opts.Parallelism
	if parallelism <= 0 {
		parallelism = runtime.NumCPU()
	}

	loadPattern := BuildLoadPattern(targetDir, opts.Recursive)

	pkgLevels, err := dir.LoadPackageLevels(ctx, loadPattern)
	if err != nil {
		return fmt.Errorf("failed to load packages: %w", err)
	}

	genOpts := generator.GeneratorOptions{
		ProcessImports: opts.Imports,
	}

	for _, pkgs := range pkgLevels {
		eg, egCtx := errgroup.WithContext(ctx)
		eg.SetLimit(parallelism)

		for _, pkg := range pkgs {
			pkg := pkg
			eg.Go(func() error {
				if opts.OnPackage != nil {
					opts.OnPackage(pkg.Pkg.Name)
				}
				return processPackage(egCtx, pkg, opts.Validate, genOpts)
			})
		}

		if err := eg.Wait(); err != nil {
			return err
		}
	}

	return nil
}

// BuildLoadPattern converts a directory path into the pattern passed to
// packages.Load. When recursive is true it appends "/...".
func BuildLoadPattern(targetDir string, recursive bool) string {
	if !recursive {
		return targetDir
	}
	cleanDir := filepath.ToSlash(filepath.Clean(targetDir))
	if cleanDir == "." {
		return "./..."
	}
	if !strings.HasPrefix(cleanDir, "/") && !strings.HasPrefix(cleanDir, "./") {
		cleanDir = "./" + cleanDir
	}
	return fmt.Sprintf("%s/...", cleanDir)
}

func processPackage(ctx context.Context, pkg *dir.Package, strict bool, opts generator.GeneratorOptions) error {
	for i, syntax := range pkg.Pkg.Syntax {
		path := pkg.Pkg.GoFiles[i]

		if strings.HasSuffix(path, "_gen.go") {
			continue
		}

		f, err := file.ProcessFile(pkg, syntax, path, strict)
		if err != nil {
			return fmt.Errorf("failed to process file %s: %w", path, err)
		}

		if f != nil && len(f.Functions) > 0 {
			if err := generator.GenerateForFile(f, opts); err != nil {
				return fmt.Errorf("failed to generate code for %s: %w", path, err)
			}
		}
	}
	return nil
}
