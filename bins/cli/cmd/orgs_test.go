package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestOrgsAPITokensCommandIsDeprecated(t *testing.T) {
	c := &cli{}
	apiTokens := c.apiTokensCmd()
	if !apiTokens.Hidden {
		t.Fatal("orgs api-tokens should be hidden")
	}

	for _, want := range []string{"create", "list", "delete"} {
		if commandByName(apiTokens, want) == nil {
			t.Fatalf("orgs api-tokens is missing subcommand %s", want)
		}
	}
}

func TestServiceAccountTokensCreateHasNameFlag(t *testing.T) {
	c := &cli{}
	tokens := commandByName(c.serviceAccountsCmd(), "tokens")
	if tokens == nil {
		t.Fatal("service-accounts tokens command missing")
	}

	create := commandByName(tokens, "create")
	if create == nil {
		t.Fatal("service-accounts tokens create command missing")
	}
	if create.Flags().Lookup("name") == nil {
		t.Fatal("service-accounts tokens create is missing --name")
	}
}

func TestServiceAccountTokensListAndDeleteFlags(t *testing.T) {
	c := &cli{}
	tokens := commandByName(c.serviceAccountsCmd(), "tokens")
	if tokens == nil {
		t.Fatal("service-accounts tokens command missing")
	}

	list := commandByName(tokens, "list")
	if list == nil {
		t.Fatal("service-accounts tokens list command missing")
	}
	if list.Flags().Lookup("search") == nil {
		t.Fatal("service-accounts tokens list is missing --search")
	}
	id := list.Flags().Lookup("id")
	if id == nil {
		t.Fatal("service-accounts tokens list is missing --id")
	}
	if _, required := id.Annotations[cobra.BashCompOneRequiredFlag]; required {
		t.Fatal("service-accounts tokens list --id should be optional")
	}

	del := commandByName(tokens, "delete")
	if del == nil {
		t.Fatal("service-accounts tokens delete command missing")
	}
	if del.Flags().Lookup("token-id") == nil {
		t.Fatal("service-accounts tokens delete is missing --token-id")
	}
	if del.Flags().Lookup("id") != nil {
		t.Fatal("service-accounts tokens delete should not have --id")
	}
}
