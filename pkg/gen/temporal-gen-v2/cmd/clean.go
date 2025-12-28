package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/config"
)

func newCleanCmd() *cobra.Command {
	cleanCmd := &cobra.Command{
		Use:   "clean [dir]",
		Short: "Clean generated files",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runCleanCmd,
	}
	cleanCmd.Flags().BoolVar(&dryRunFlag, "dry-run", false, "Dry run for clean command")
	cleanCmd.Flags().BoolVarP(&recursiveFlag, "recursive", "r", false, "Recursively process subdirectories")
	return cleanCmd
}

func runCleanCmd(cmd *cobra.Command, args []string) error {
	targetDir := getDir(args)
	return runClean(targetDir, dryRunFlag, recursiveFlag)
}

func runClean(dir string, dryRun bool, recursive bool) error {
	fmt.Printf("Cleaning generated files in %s (recursive=%v)...\n", dir, recursive)

	// If not recursive, we just ReadDir
	if !recursive {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if strings.HasSuffix(entry.Name(), "_gen.go") {
				path := filepath.Join(dir, entry.Name())
				if err := checkAndRemove(path, dryRun); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Recursive
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, "_gen.go") {
			return checkAndRemove(path, dryRun)
		}
		return nil
	})
}

func checkAndRemove(path string, dryRun bool) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if strings.Contains(string(content), config.Watermark) {
		if dryRun {
			fmt.Printf("Would delete: %s\n", path)
		} else {
			fmt.Printf("Deleting: %s\n", path)
			if err := os.Remove(path); err != nil {
				return err
			}
		}
	}
	return nil
}
