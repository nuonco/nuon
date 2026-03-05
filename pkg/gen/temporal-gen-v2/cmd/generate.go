package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

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
	return generateCmd
}

func runGen(cmd *cobra.Command, args []string) error {
	targetDir := getDir(args)
	strict := validateFlag

	// Create generator options from flags
	opts := generator.GeneratorOptions{
		ProcessImports: importsFlag,
	}

	if cleanupFlag {
		if err := runClean(targetDir, false, recursiveFlag); err != nil {
			return fmt.Errorf("failed to cleanup: %w", err)
		}
	}

	loadDir := targetDir
	if recursiveFlag {
		cleanDir := filepath.ToSlash(filepath.Clean(targetDir))
		if cleanDir == "." {
			loadDir = "./..."
		} else {
			if !strings.HasPrefix(cleanDir, "/") && !strings.HasPrefix(cleanDir, "./") {
				cleanDir = "./" + cleanDir
			}
			loadDir = fmt.Sprintf("%s/...", cleanDir)
		}
	}

	fmt.Printf("Running %s generator in %s...\n", config.AnnotationPrefix, loadDir)

	ctx := context.Background()
	pkgPaths, err := dir.GetDependencyOrder(ctx, loadDir)
	if err != nil {
		return fmt.Errorf("failed to resolve package order: %w", err)
	}
	fmt.Printf("Identified %d packages to process\n", len(pkgPaths))

	for _, pkgPath := range pkgPaths {
		pkg, err := dir.LoadPackage(ctx, pkgPath)
		if err != nil {
			return fmt.Errorf("failed to load package %s: %w", pkgPath, err)
		}

		fmt.Printf("Processing package %s\n", pkg.Pkg.Name)
		for i, syntax := range pkg.Pkg.Syntax {
			path := pkg.Pkg.GoFiles[i]

			// Skip generated files
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
	}

	return nil
}
