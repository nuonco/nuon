package cmd

import (
	"context"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/config"
	temporalgen "github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/lib"
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

func runGen(cmd *cobra.Command, args []string) error {
	targetDir := getDir(args)

	loadPattern := temporalgen.BuildLoadPattern(targetDir, recursiveFlag)
	fmt.Printf("Running %s generator in %s...\n", config.AnnotationPrefix, loadPattern)

	ctx := context.Background()
	return temporalgen.Generate(ctx, temporalgen.Options{
		Dir:         targetDir,
		Recursive:   recursiveFlag,
		Cleanup:     cleanupFlag,
		Validate:    validateFlag,
		Imports:     importsFlag,
		Parallelism: parallelismFlag,
		OnPackage: func(name string) {
			fmt.Printf("  processing %s\n", name)
		},
	})
}
