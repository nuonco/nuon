package cmd

import (
	"github.com/spf13/cobra"
)

func (c *cli) debugCmd() *cobra.Command {
	debug := &cobra.Command{
		Use:    "debug",
		Short:  "Debug helpers (requires NUON_DEBUG=true)",
		Hidden: true,
	}

	debug.AddCommand(&cobra.Command{
		Use:         "noop",
		Short:       "Do nothing, skipping all initialization",
		Annotations: annotations(skipAuthAnnotation(), outputsAnnotation(OutputTable)),
		Run:         func(*cobra.Command, []string) {},
	})

	debug.AddCommand(&cobra.Command{
		Use:               "noop-init",
		Short:             "Do nothing, after init but without auth",
		PersistentPreRunE: c.persistentPreRunE,
		Annotations:       annotations(skipAuthAnnotation(), outputsAnnotation(OutputTable)),
		Run:               func(*cobra.Command, []string) {},
	})

	debug.AddCommand(&cobra.Command{
		Use:               "noop-auth",
		Short:             "Do nothing, after init with auth",
		PersistentPreRunE: c.persistentPreRunE,
		Annotations:       outputsAnnotation(OutputTable),
		Run:               func(*cobra.Command, []string) {},
	})

	return debug
}
