package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestWorkerProviderGraphs(t *testing.T) {
	for _, ns := range []string{
		"all", "general", "orgs", "apps", "components", "installs", "install-crons",
		"runners", "runner-healthcheck-crons", "actions", "onboardings", "vcs",
	} {
		t.Run(ns, func(t *testing.T) {
			prev := namespace
			namespace = ns
			t.Cleanup(func() { namespace = prev })

			providers := append((&cli{}).workerProviders(), fx.NopLogger)
			require.NoError(t, fx.ValidateApp(providers...))
		})
	}
}
