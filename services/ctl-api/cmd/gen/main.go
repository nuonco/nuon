package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

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
		"run", "github.com/swaggo/swag/cmd/swag@latest",
		"init",
		"--instanceName", "runner",
		"--output", "docs/runner",
		"--parseDependency",
		"--parseInternal", "-g", "runner.go",
		"--markdownFiles", "docs/runner/descriptions",
		"-t", "orgs/runner,apps/runner,general/runner,sandboxes/runner,installs/runner,installers/runner,components/runner,runners/runner,runners/auth,actions/runner",
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
		"run", "github.com/swaggo/swag/cmd/swag@latest",
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
		"run", "github.com/swaggo/swag/cmd/swag@latest",
		"init",
		"--parseDependency",
		"--output", "docs/public",
		"--parseInternal",
		"-g", "public.go",
		"--markdownFiles", "docs/public/descriptions",
		"-t", "auth,accounts,apps,actions,components,installs,installers,general,orgs,releases,sandboxes,vcs,runners",
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
	targetsFlag := flag.String("targets", "", "comma-separated targets: public,public-v3,runner,admin,temporal")
	flag.Parse()

	targets := parseTargets(*targetsFlag)
	if targets.isEmpty() {
		log.Fatal("no generation targets specified")
	}
	if targets.has("public-v3") {
		targets.add("public")
	}

	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	// Phase 1: Run independent tasks in parallel
	recorder := newTimingRecorder()
	eg, ctx := errgroup.WithContext(ctx)
	if targets.has("runner") {
		eg.Go(func() error {
			return recorder.time("Runner schema", func() error {
				return generateRunnerSchema(ctx)
			})
		})
	}
	if targets.has("admin") {
		eg.Go(func() error {
			return recorder.time("Admin schema", func() error {
				return generateAdminSchema(ctx)
			})
		})
	}
	if targets.has("temporal") {
		eg.Go(func() error {
			return recorder.time("Temporal", func() error {
				return runTemporalGen(ctx)
			})
		})
	}

	if targets.has("public") {
		eg.Go(func() error {
			if err := recorder.time("Public schema", func() error {
				return generatePublicSchema(ctx)
			}); err != nil {
				return err
			}
			if targets.has("public-v3") {
				return recorder.time("Public v3", func() error {
					return generatePublicOAPI3Spec(ctx)
				})
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}

	recorder.printSummary(os.Stdout)
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

type targetSet map[string]struct{}

func parseTargets(flagValue string) targetSet {
	value := strings.TrimSpace(flagValue)
	if value == "" {
		value = strings.TrimSpace(os.Getenv("NUON_GEN_TARGETS"))
	}

	set := targetSet{}
	if value == "" {
		for _, target := range []string{"public", "public-v3", "runner", "admin", "temporal"} {
			set.add(target)
		}
		return set
	}

	for _, part := range strings.Split(value, ",") {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		if trimmed == "sdk" {
			set.add("public")
			set.add("runner")
			continue
		}
		set.add(trimmed)
	}

	return set
}

func (t targetSet) add(target string) {
	if t != nil {
		t[target] = struct{}{}
	}
}

func (t targetSet) has(target string) bool {
	_, ok := t[target]
	return ok
}

func (t targetSet) isEmpty() bool {
	return len(t) == 0
}

func (t targetSet) String() string {
	keys := make([]string, 0, len(t))
	for key := range t {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

type timingRecorder struct {
	mu      sync.Mutex
	timings map[string]time.Duration
}

func newTimingRecorder() *timingRecorder {
	return &timingRecorder{timings: make(map[string]time.Duration)}
}

func (r *timingRecorder) time(name string, fn func() error) error {
	start := time.Now()
	err := fn()
	r.mu.Lock()
	r.timings[name] = time.Since(start)
	r.mu.Unlock()
	return err
}

func (r *timingRecorder) printSummary(out *os.File) {
	entries := make([]timingEntry, 0, len(r.timings))
	for name, duration := range r.timings {
		entries = append(entries, timingEntry{name: name, duration: duration})
	}
	if len(entries) == 0 {
		return
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].duration > entries[j].duration
	})

	fmt.Fprintln(out, "")
	fmt.Fprintf(out, "%-22s | %10s\n", "Step", "Seconds")
	fmt.Fprintf(out, "%-22s-+-%10s\n", "----------------------", "----------")
	for _, entry := range entries {
		fmt.Fprintf(out, "%-22s | %10.1f\n", entry.name, entry.duration.Seconds())
	}

	longest := entries[0]
	fmt.Fprintf(out, "\nLongest: %s (%.1fs)\n", longest.name, longest.duration.Seconds())
}

type timingEntry struct {
	name     string
	duration time.Duration
}
