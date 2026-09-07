package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestBranchesCommandIsTopLevelAndDeprecatedAlias(t *testing.T) {
	c := &cli{}
	root := c.branchesCmd()
	if root.Deprecated != "" {
		t.Fatalf("canonical command should not be deprecated, got %q", root.Deprecated)
	}
	if root.GroupID != CoreGroup.ID {
		t.Fatalf("expected core group, got %q", root.GroupID)
	}

	alias := c.newBranchesCmd(true)
	if alias.Deprecated == "" {
		t.Fatal("apps branches alias should be deprecated")
	}

	names := map[string]bool{}
	for _, child := range root.Commands() {
		names[child.Name()] = true
	}
	for _, want := range []string{"list", "get", "create", "sync", "trigger", "preview", "delete", "runs"} {
		if !names[want] {
			t.Fatalf("missing subcommand %s", want)
		}
	}

	syncCmd := commandByName(root, "sync")
	if syncCmd == nil {
		t.Fatal("sync command missing")
	}
	if syncCmd.Flags().Lookup("file") == nil || syncCmd.Flags().Lookup("app-id") == nil {
		t.Fatal("sync is missing --file or --app-id")
	}
	if syncCmd.Flags().Lookup("app-id").Shorthand != "a" {
		t.Fatal("--app-id shorthand should be -a")
	}
}

func commandByName(parent *cobra.Command, name string) *cobra.Command {
	for _, child := range parent.Commands() {
		if child.Name() == name {
			return child
		}
	}
	return nil
}
