package cmd

import (
	"testing"

	"go.uber.org/fx"
)

func TestRunnerCommandDependencyGraphs(t *testing.T) {
	c := new(cli)
	for name, options := range map[string][]fx.Option{
		"build": c.buildOptions(),
		"run":   c.runOptions(),
	} {
		t.Run(name, func(t *testing.T) {
			if err := fx.ValidateApp(options...); err != nil {
				t.Fatal(err)
			}
		})
	}
}
