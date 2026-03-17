package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/sync/errgroup"

	"github.com/go-playground/validator/v10"

	"github.com/a-h/templ/cmd/templ/generatecmd"
	"github.com/nuonco/nuon/pkg/command"
	temporalgen "github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/lib"
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
		"-t", "orgs/admin,actions/admin,apps/admin,general/admin,sandboxes/admin,installs/admin,installers/admin,components/admin,runners/admin,auth/admin,queues/admin",
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

func generateTemporal(ctx context.Context) error {
	return temporalgen.Generate(ctx, temporalgen.Options{
		Dir:         ".",
		Recursive:   true,
		Cleanup:     true,
		Validate:    true,
		Imports:     true,
		Parallelism: runtime.NumCPU(),
	})
}

func generateTempl(ctx context.Context) error {
	return generatecmd.Run(ctx, os.Stdout, os.Stderr, []string{"-path", "./internal/app/admin-dashboard"})
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
	if targets.has("temporal") {
		eg.Go(func() error {
			return recorder.time("Temporal gen", func() error {
				return generateTemporal(ctx)
			})
		})
	}
	if targets.has("templ") {
		eg.Go(func() error {
			return recorder.time("Templ gen", func() error {
				return generateTempl(ctx)
			})
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
		for _, target := range []string{"public", "public-v3", "runner", "admin", "temporal", "templ"} {
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

	t := table.NewWriter()
	t.SetOutputMirror(out)
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = false
	t.AppendHeader(table.Row{"Step", "Seconds"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignLeft},
		{Number: 2, Align: text.AlignRight},
	})

	for _, entry := range entries {
		t.AppendRow(table.Row{entry.name, fmt.Sprintf("%.1f", entry.duration.Seconds())})
	}

	fmt.Fprintln(out, "")
	t.Render()

	longest := entries[0]
	fmt.Fprintf(out, "\nLongest: %s (%.1fs)\n", longest.name, longest.duration.Seconds())
}

type timingEntry struct {
	name     string
	duration time.Duration
}
