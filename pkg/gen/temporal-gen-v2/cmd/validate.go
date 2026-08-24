package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/config"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/dir"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/file"
	temporalgen "github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/lib"
)

func newValidateCmd() *cobra.Command {
	validateCmd := &cobra.Command{
		Use:   "validate [dir]",
		Short: "Validate annotations",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runValidateCmd,
	}
	validateCmd.Flags().BoolVarP(&recursiveFlag, "recursive", "r", false, "Recursively process subdirectories")
	validateCmd.Flags().StringVar(&configFlag, "config", "", "Path to temporal-gen.yaml (default: discovered by walking up from [dir])")
	validateCmd.Flags().BoolVar(&noConfigFlag, "no-config", false, "Skip label config discovery entirely")
	return validateCmd
}

func runValidateCmd(cmd *cobra.Command, args []string) error {
	targetDir := getDir(args)
	return runValidate(targetDir, recursiveFlag)
}

func runValidate(targetDir string, recursive bool) error {
	loadDir := targetDir
	if recursive {
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

	// Load the label config up front so a malformed temporal-gen.yaml fails
	// validation even when nothing references @labels yet.
	labelCfg, err := temporalgen.ResolveConfig(temporalgen.Options{
		Dir:        targetDir,
		ConfigPath: configFlag,
		NoConfig:   noConfigFlag,
	})
	if err != nil {
		return err
	}
	if labelCfg != nil {
		fmt.Printf("using label config %s\n", labelCfg.Path())
	}

	fmt.Printf("Validating %s annotations in %s...\n", config.AnnotationPrefix, loadDir)
	ctx := context.Background()
	pkgs, err := dir.LoadPackages(ctx, loadDir)
	if err != nil {
		return fmt.Errorf("failed to load packages: %w", err)
	}

	hasError := false
	for _, pkg := range pkgs {
		for i, syntax := range pkg.Pkg.Syntax {
			path := pkg.Pkg.GoFiles[i]

			// Skip generated files
			if strings.HasSuffix(path, "_gen.go") {
				continue
			}

			if _, err := file.ProcessFile(pkg, syntax, path, true, labelCfg); err != nil {
				fmt.Fprintf(os.Stderr, "Validation error in %s: %v\n", path, err)
				hasError = true
			}
		}
	}

	if hasError {
		return fmt.Errorf("validation failed")
	}

	return nil
}
