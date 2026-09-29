package temporalgen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/config"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/dir"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/file"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/generator"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/tags"
)

type Options struct {
	Dir string

	Recursive bool

	Cleanup bool

	Validate bool

	Imports bool

	Parallelism int

	OnPackage func(pkgName string)

	Tags *tags.Config

	ConfigPath string

	NoConfig bool
}

func ResolveConfig(opts Options) (*tags.Config, error) {
	if opts.NoConfig {
		return nil, nil
	}
	if opts.Tags != nil {
		if err := opts.Tags.Validate(); err != nil {
			return nil, err
		}
		return opts.Tags, nil
	}
	if opts.ConfigPath != "" {
		return tags.Load(opts.ConfigPath)
	}
	targetDir := opts.Dir
	if targetDir == "" {
		targetDir = "."
	}
	return tags.Discover(targetDir)
}

func Generate(ctx context.Context, opts Options) error {
	targetDir := opts.Dir
	if targetDir == "" {
		targetDir = "."
	}

	parallelism := opts.Parallelism
	if parallelism <= 0 {
		parallelism = runtime.NumCPU()
	}

	tagCfg, err := ResolveConfig(opts)
	if err != nil {
		return err
	}
	if tagCfg != nil && tagCfg.Path() != "" {
		fmt.Printf("using tag config %s\n", tagCfg.Path())
	}

	if opts.Cleanup {
		if err := Clean(targetDir, opts.Recursive); err != nil {
			return fmt.Errorf("failed to cleanup: %w", err)
		}
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
				return processPackage(egCtx, pkg, opts.Validate, genOpts, tagCfg)
			})
		}

		if err := eg.Wait(); err != nil {
			return err
		}
	}

	return nil
}

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

func Clean(dir string, recursive bool) error {
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
				if err := removeIfGenerated(path); err != nil {
					return err
				}
			}
		}
		return nil
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, "_gen.go") {
			return removeIfGenerated(path)
		}
		return nil
	})
}

func removeIfGenerated(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if strings.Contains(string(content), config.Watermark) {
		return os.Remove(path)
	}
	return nil
}

func processPackage(ctx context.Context, pkg *dir.Package, strict bool, opts generator.GeneratorOptions, cfg *tags.Config) error {
	for i, syntax := range pkg.Pkg.Syntax {
		path := pkg.Pkg.GoFiles[i]

		if strings.HasSuffix(path, "_gen.go") {
			continue
		}

		f, err := file.ProcessFile(pkg, syntax, path, strict, cfg)
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
