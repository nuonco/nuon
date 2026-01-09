package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"golang.org/x/sync/errgroup"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/command"
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
		"-t", "orgs/admin,actions/admin,apps/admin,general/admin,sandboxes/admin,installs/admin,installers/admin,components/admin,runners/admin,auth/admin",
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
		"-t", "accounts,apps,actions,components,installs,installers,general,orgs,releases,sandboxes,vcs,runners",
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

func generatePublicOAPI3Spec(ctx context.Context) error {
	// Load the generated Swagger 2.0 spec
	doc, err := LoadPublicOAPI2Spec()
	if err != nil {
		return fmt.Errorf("unable to load swagger spec: %w", err)
	}

	// Convert to OpenAPI 3.0
	oapi3Doc, err := openapi2conv.ToV3(doc)
	if err != nil {
		return fmt.Errorf("unable to convert to openapi v3: %w", err)
	}

	// Write to docs/public/swagger-v3.json
	outputPath := "docs/public/swagger-v3.json"
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("unable to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(oapi3Doc); err != nil {
		return fmt.Errorf("unable to write json: %w", err)
	}

	fmt.Fprintf(os.Stdout, "✅ successfully generated public openapi v3 spec: %s\n", outputPath)
	return nil
}

func runTemporalGen(ctx context.Context) error {
	// Build a binary to reuse per-directory
	binpath, err := compileToTemp(ctx, "github.com/nuonco/nuon/pkg/gen/temporal-gen")
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

	// Phase 1: Run independent tasks in parallel
	eg, ctx := errgroup.WithContext(ctx)
	parallelFns := []func(context.Context) error{
		generateRunnerSchema,
		generateAdminSchema,
		runTemporalGen,
	}

	for _, fn := range parallelFns {
		eg.Go(func() error {
			return fn(ctx)
		})
	}

	// Phase 2: Generate public schema first, then convert to v3
	eg.Go(func() error {
		if err := generatePublicSchema(ctx); err != nil {
			return err
		}
		return generatePublicOAPI3Spec(ctx)
	})

	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}

func LoadPublicOAPI2Spec() (*openapi2.T, error) {
	byts, err := os.ReadFile("docs/public/swagger.json")
	if err != nil {
		return nil, fmt.Errorf("unable to read swagger.json file: %w", err)
	}

	var doc openapi2.T
	err = json.Unmarshal(byts, &doc)
	if err != nil {
		return nil, fmt.Errorf("unable to convert open api spec to json: %w", err)
	}

	return &doc, nil
}
