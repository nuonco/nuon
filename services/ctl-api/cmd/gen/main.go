package main

import (
	"context"
	"fmt"
	"log"
	"os"

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

func main() {
	ctx := context.Background()
	ctx, cancelFn := context.WithCancel(ctx)
	defer cancelFn()

	eg, ctx := errgroup.WithContext(ctx)
	fns := []func(context.Context) error{
		generateRunnerSchema,
		generatePublicSchema,
		generateAdminSchema,
	}

	for idx, fn := range fns {
		if err := fn(ctx); err != nil {
			log.Fatal("failed on %d %s", idx, err.Error())
		}

		// NOTE(jm): this is not parallelized any more because docker builds seem to run out of resources when
		// building. Since it's not a huge speedup, I'm just disabling it for now, and we can revisit increasing
		// memory during builds later.
		continue
		eg.Go(func() error {
			return fn(ctx)
		})
	}
	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}
