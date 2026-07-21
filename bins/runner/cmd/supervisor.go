package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/pkg/actions/supervisor"
)

func (c *cli) registerSupervisor() error {
	cmd := &cobra.Command{
		Use:    "actions-supervisor",
		Long:   "run an image-backed action step inside its container (mounted by the mng runner as the container entrypoint)",
		Hidden: true,
		Run:    c.runSupervisor,
	}
	cmd.Flags().String("script", "", "path to the rendered action step script")
	cmd.Flags().String("workdir", "", "working directory to execute the script in")
	rootCmd.AddCommand(cmd)
	return nil
}

func (c *cli) runSupervisor(cmd *cobra.Command, _ []string) {
	script, _ := cmd.Flags().GetString("script")
	workdir, _ := cmd.Flags().GetString("workdir")

	if script == "" {
		fmt.Fprintln(os.Stderr, "actions-supervisor: --script is required")
		os.Exit(1)
	}
	if workdir == "" {
		workdir = "."
	}

	exitCode, err := supervisor.Run(cmd.Context(), script, workdir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "actions-supervisor: %v\n", err)
		os.Exit(1)
	}
	os.Exit(exitCode)
}
