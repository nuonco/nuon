package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/config"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/dir"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/file"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/generator"
)

func newGenerateCmd() *cobra.Command {
	generateCmd := &cobra.Command{
		Use:   "generate [dir]",
		Short: "Generate code",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runGen,
	}
	generateCmd.Flags().BoolVar(&validateFlag, "validate", false, "Fail on validation errors")
	generateCmd.Flags().BoolVar(&cleanupFlag, "cleanup", false, "Cleanup generated files before generating")
	generateCmd.Flags().BoolVarP(&recursiveFlag, "recursive", "r", false, "Recursively process subdirectories")
	generateCmd.Flags().BoolVar(&importsFlag, "imports", false, "Process imports using golang.org/x/tools/imports library")
	generateCmd.Flags().IntVarP(&parallelismFlag, "parallelism", "p", runtime.NumCPU(), "Number of packages to process concurrently per dependency level")
	return generateCmd
}

func processPackage(pkg *dir.Package, strict bool, opts generator.GeneratorOptions) error {
	fmt.Printf("  Processing package %s\n", pkg.Pkg.Name)
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

func runGen(cmd *cobra.Command, args []string) error {
	targetDir := getDir(args)
	strict := validateFlag

	opts := generator.GeneratorOptions{
		ProcessImports: importsFlag,
	}

	if cleanupFlag {
		if err := runClean(targetDir, false, recursiveFlag); err != nil {
			return fmt.Errorf("failed to cleanup: %w", err)
		}
	}

	loadPattern := targetDir
	if recursiveFlag {
		cleanDir := filepath.ToSlash(filepath.Clean(targetDir))
		if cleanDir == "." {
			loadPattern = "./..."
		} else {
			if !strings.HasPrefix(cleanDir, "/") && !strings.HasPrefix(cleanDir, "./") {
				cleanDir = "./" + cleanDir
			}
			loadPattern = fmt.Sprintf("%s/...", cleanDir)
		}
	}

	fmt.Printf("Running %s generator in %s...\n", config.AnnotationPrefix, loadPattern)

	ctx := context.Background()

	// Load all packages in a single packages.Load call, grouped by dependency level.
	// This is the critical path: one toolchain invocation instead of one per package.
	pkgLevels, err := dir.LoadPackageLevels(ctx, loadPattern)
	if err != nil {
		return fmt.Errorf("failed to load packages: %w", err)
	}

	totalPkgs := 0
	for _, lvl := range pkgLevels {
		totalPkgs += len(lvl)
	}
	fmt.Printf("Identified %d packages across %d dependency levels (parallelism=%d)\n",
		totalPkgs, len(pkgLevels), parallelismFlag)

	for levelIdx, pkgs := range pkgLevels {
		fmt.Printf("Processing level %d/%d (%d packages)...\n", levelIdx+1, len(pkgLevels), len(pkgs))

		eg, _ := errgroup.WithContext(ctx)
		eg.SetLimit(parallelismFlag)

		for _, pkg := range pkgs {
			pkg := pkg
			eg.Go(func() error {
				return processPackage(pkg, strict, opts)
			})
		}

		if err := eg.Wait(); err != nil {
			return err
		}
	}

	return nil
}
