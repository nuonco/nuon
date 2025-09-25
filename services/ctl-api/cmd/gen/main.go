package main

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/pkg/command"
)

var v *validator.Validate

func init() {
	v = validator.New()
}

func generateRunnerSchema(ctx context.Context) error {
	args := []string{
		"run", "github.com/swaggo/swag/cmd/swag",
		"init",
		"--instanceName", "runner",
		"--output", "docs/runner",
		"--parseDependency",
		"--parseInternal", "-g", "runner.go",
		"--markdownFiles", "docs/runner/descriptions",
		"-t", "orgs/runner,apps/runner,general/runner,sandboxes/runner,installs/runner,installers/runner,components/runner,runners/runner,actions/runner",
	}

	cmd, err := command.New(v,
		command.WithInheritedEnv(),
		command.WithCmd("go"),
		command.WithArgs(args),
		command.WithLinePrefix("runner-schema"),
	)
	if err != nil {
		return fmt.Errorf("unable to create command: %w", err)
	}

	if err := cmd.Exec(ctx); err != nil {
		return fmt.Errorf("unable to execute command: %w", err)
	}

	fmt.Fprintf(os.Stdout, "✅ successfully generated runner schema\n")
	return nil
}

func generateAdminSchema(ctx context.Context) error {
	args := []string{
		"run", "github.com/swaggo/swag/cmd/swag",
		"init",
		"--instanceName", "admin",
		"--output", "docs/admin",
		"--parseDependency",
		"--parseInternal",
		"-g", "admin.go",
		"--markdownFiles", "docs/admin/descriptions",
		"-t", "orgs/admin,actions/admin,apps/admin,general/admin,sandboxes/admin,installs/admin,installers/admin,components/admin,runners/admin",
	}

	cmd, err := command.New(v,
		command.WithInheritedEnv(),
		command.WithCmd("go"),
		command.WithArgs(args),
		command.WithLinePrefix("admin-schema"),
	)
	if err != nil {
		return fmt.Errorf("unable to create command: %w", err)
	}

	if err := cmd.Exec(ctx); err != nil {
		return fmt.Errorf("unable to execute command: %w", err)
	}

	fmt.Fprintf(os.Stdout, "✅ successfully generated admin schema\n")
	return nil
}

func generatePublicSchema(ctx context.Context) error {
	args := []string{
		"run", "github.com/swaggo/swag/cmd/swag",
		"init",
		"--parseDependency",
		"--output", "docs/public",
		"--parseInternal",
		"-g", "public.go",
		"--markdownFiles", "docs/public/descriptions",
		"-t", "apps,actions,components,installs,installers,general,orgs,releases,sandboxes,vcs,runners",
	}

	cmd, err := command.New(v,
		command.WithInheritedEnv(),
		command.WithCmd("go"),
		command.WithArgs(args),
		command.WithLinePrefix("public-schema"),
	)
	if err != nil {
		return fmt.Errorf("unable to create command: %w", err)
	}

	if err := cmd.Exec(ctx); err != nil {
		return fmt.Errorf("unable to execute command: %w", err)
	}

	fmt.Fprintf(os.Stdout, "✅ successfully generated public schema\n")
	return nil
}

func runTemporalGen(ctx context.Context) error {
	// Build a binary to reuse per-directory
	binpath, err := compileToTemp(ctx, "github.com/powertoolsdev/mono/pkg/gen/temporal-gen")
	if err != nil {
		return fmt.Errorf("unable to compile temporal-gen binary: %w", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to get current working directory: %w", err)
	}

	paths := make(chan string)
	var pathmap sync.Map
	eg, _ := errgroup.WithContext(ctx)
	numWorkers := runtime.NumCPU()
	var inerr error

	for i := 0; i < numWorkers; i++ {
		eg.Go(func() error {
			for path := range paths {
				dir := filepath.Dir(path)
				if _, has := pathmap.Load(dir); has {
					continue
				}
				byt, err := os.ReadFile(path)
				if err != nil {
					return fmt.Errorf("unable to read file %s: %w", path, err)
				}

				if bytes.Contains(byt, []byte("\n// @temporal-gen ")) {
					pathmap.Store(dir, struct{}{})
				}
			}

			return nil
		})
	}

	err = filepath.WalkDir(wd, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Type().IsRegular() {
			paths <- path
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("unable to walk directory: %w", err)
	}

	close(paths)
	eg.Wait()
	if inerr != nil {
		return inerr
	}

	eg, _ = errgroup.WithContext(ctx)
	dirs := make(chan string)
	for i := 0; i < numWorkers; i++ {
		eg.Go(func() error {
			for dir := range dirs {
				cmd, err := command.New(v,
					command.WithInheritedEnv(),
					command.WithCmd(binpath),
					command.WithCwd(dir),
				)
				if err != nil {
					inerr = fmt.Errorf("unable to create command: %w", err)
					continue
				}
				if err := cmd.Exec(ctx); err != nil {
					inerr = fmt.Errorf("error running temporal-gen on %s: %w", dir, err)
				}
			}

			return nil
		})
	}

	pathmap.Range(func(k, _ any) bool {
		dirs <- k.(string)
		return true
	})

	close(dirs)
	eg.Wait()

	return inerr
}

func compileToTemp(ctx context.Context, path string) (string, error) {
	// Compile the temporal-gen binary for the given path
	// This is a placeholder function and should be implemented as needed
	name := filepath.Base(path)

	tmpdir, err := os.MkdirTemp(os.TempDir(), name)
	if err != nil {
		return "", fmt.Errorf("unable to create temporary directory: %w", err)
	}

	binpath := filepath.Join(tmpdir, name)

	args := []string{
		"build",
		"-o", binpath,
		path,
	}

	cmd, err := command.New(v,
		command.WithInheritedEnv(),
		command.WithCmd("go"),
		command.WithArgs(args),
	)
	if err != nil {
		return "", fmt.Errorf("unable to create command: %w", err)
	}
	if err := cmd.Exec(ctx); err != nil {
		return "", fmt.Errorf("unable to execute command: %w", err)
	}
	return binpath, nil
}

func main() {
	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	eg, ctx := errgroup.WithContext(ctx)
	fns := []func(context.Context) error{
		generateRunnerSchema,
		generatePublicSchema,
		generateAdminSchema,
		runTemporalGen,
	}

	for _, fn := range fns {
		eg.Go(func() error {
			return fn(ctx)
		})
	}
	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}
