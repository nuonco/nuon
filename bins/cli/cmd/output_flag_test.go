package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestOutputShorthandParsesOnCommandsThatPaginate(t *testing.T) {
	root := (&cli{}).rootCmd()
	list, _, err := root.Find([]string{"builds", "list"})
	if err != nil {
		t.Fatal(err)
	}
	if err := list.ParseFlags([]string{"-o", "json"}); err != nil {
		t.Fatal(err)
	}
	if Output != "json" {
		t.Fatalf("-o json set output to %q", Output)
	}

	var walk func(*cobra.Command)
	walk = func(cmd *cobra.Command) {
		t.Helper()
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("%s: %v", cmd.CommandPath(), recovered)
			}
		}()
		if err := cmd.ParseFlags(nil); err != nil {
			t.Fatalf("%s: %v", cmd.CommandPath(), err)
		}
		for _, child := range cmd.Commands() {
			walk(child)
		}
	}
	walk(root)
}
