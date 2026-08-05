package cmd

import (
	"github.com/spf13/cobra"
)

// debugCmd holds no-op commands for separating fixed CLI overhead from API
// latency when profiling. Registered only when NUON_DEBUG=true.
func (c *cli) debugCmd() *cobra.Command {
	debug := &cobra.Command{
		Use:    "debug",
		Short:  "Debug helpers (requires NUON_DEBUG=true)",
		Hidden: true,
	}

	// No PersistentPreRunE: process start and command-tree construction only.
	debug.AddCommand(&cobra.Command{
		Use:         "noop",
		Short:       "Do nothing, skipping all initialization",
		Annotations: annotations(skipAuthAnnotation(), outputsAnnotation(OutputTable)),
		Run:         func(*cobra.Command, []string) {},
	})

	// Adds config, API client, sentry and analytics init, but no auth.
	debug.AddCommand(&cobra.Command{
		Use:               "noop-init",
		Short:             "Do nothing, after init but without auth",
		PersistentPreRunE: c.persistentPreRunE,
		Annotations:       annotations(skipAuthAnnotation(), outputsAnnotation(OutputTable)),
		Run:               func(*cobra.Command, []string) {},
	})

	// Adds initUser, which is one API round trip.
	debug.AddCommand(&cobra.Command{
		Use:               "noop-auth",
		Short:             "Do nothing, after init with auth",
		PersistentPreRunE: c.persistentPreRunE,
		Annotations:       outputsAnnotation(OutputTable),
		Run:               func(*cobra.Command, []string) {},
	})

	return debug
}
