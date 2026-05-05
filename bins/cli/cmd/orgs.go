package cmd

import (
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/services/orgs"
)

func (c *cli) orgsCmd() *cobra.Command {
	var (
		id           string
		name         string
		sandbox      bool
		offset       int
		limit        int
		search       string
		email        string
		noSelect     bool
		connectionID string
	)

	orgsCmd := &cobra.Command{
		Use:               "orgs",
		Short:             "Manage your organizations",
		Aliases:           []string{"a"},
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           AdditionalGroup.ID,
	}

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get current org",
		Long:  "Get the org you are currently authenticated with",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.Current(cmd.Context(), PrintJSON)
		}),
	}
	orgsCmd.AddCommand(getCmd)

	currentCmd := &cobra.Command{
		Use:   "current",
		Short: "Get current org (deprecated)",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			printDeprecatedCommandWarning(cmd, "Use `nuon orgs get` instead")

			svc := orgs.New(c.apiClient, c.cfg)
			return svc.Current(cmd.Context(), PrintJSON)
		}),
	}
	currentCmd.Hidden = true
	orgsCmd.AddCommand(currentCmd)

	apiTokenCmd := &cobra.Command{
		Use:   "api-token",
		Short: "Get api token",
		Long:  "Get api token that is active for current org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.APIToken(cmd.Context(), PrintJSON)
		}),
	}
	orgsCmd.AddCommand(apiTokenCmd)

	idCmd := &cobra.Command{
		Use:   "id",
		Short: "Get current org id",
		Long:  "Get id for current org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.ID(cmd.Context(), PrintJSON)
		}),
	}
	orgsCmd.AddCommand(idCmd)

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List orgs",
		Long:    "List all your orgs",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.List(cmd.Context(), offset, limit, search, PrintJSON)
		}),
	}
	listCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Limit for pagination")
	listCmd.Flags().StringVar(&search, "search", "", "Search orgs by name")
	orgsCmd.AddCommand(listCmd)

	listVCSConnections := &cobra.Command{
		Use:   "list-vcs-connections",
		Short: "List VCS connections",
		Long:  "List all connected GitHub accounts",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.VCSConnections(cmd.Context(), offset, limit, PrintJSON)
		}),
	}
	listVCSConnections.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listVCSConnections.Flags().IntVarP(&limit, "limit", "l", 20, "Limit for pagination")
	orgsCmd.AddCommand(listVCSConnections)

	deleteVCSConnectionCmd := &cobra.Command{
		Use:   "delete-vcs-connection",
		Short: "Delete VCS Connection",
		Long:  "Delete a VCS connection from your Nuon org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.DeleteVCSConnection(cmd.Context(), connectionID, PrintJSON)
		}),
	}
	deleteVCSConnectionCmd.MarkFlagRequired("org-id")
	deleteVCSConnectionCmd.Flags().StringVar(&connectionID, "connection-id", "", "The VCS Connection ID you want to remove")
	deleteVCSConnectionCmd.MarkFlagRequired("connection-id")
	orgsCmd.AddCommand(deleteVCSConnectionCmd)

	connectGithubCmd := &cobra.Command{
		Use:   "connect-github",
		Short: "Connect GitHub account",
		Long:  "Connect GitHub account to your Nuon org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.ConnectGithub(cmd.Context())
		}),
	}
	connectGithubCmd.MarkFlagRequired("org-id")
	orgsCmd.AddCommand(connectGithubCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create new org",
		Long:  "Create a new org and set it as the current org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.Create(cmd.Context(), name, sandbox, noSelect, PrintJSON)
		}),
	}
	createCmd.Flags().StringVarP(&name, "name", "n", "", "The name of your new org")
	createCmd.MarkFlagRequired("name")
	createCmd.Flags().BoolVar(&sandbox, "sandbox-mode", false, "Create org in sandbox mode")
	createCmd.Flags().BoolVar(&noSelect, "no-select", false, "Do not automatically set the new org as the current org")
	orgsCmd.AddCommand(createCmd)

	selectOrgCmd := &cobra.Command{
		Use:         "select",
		Short:       "Select your current org",
		Long:        "Select your current org from a list or by org ID",
		Annotations: tuiAnnotation(TUIContextual),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.Select(cmd.Context(), id, offset, limit, PrintJSON)
		}),
	}
	selectOrgCmd.Flags().StringVar(&id, "org", "", "The ID of the org you want to use")
	selectOrgCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	selectOrgCmd.Flags().IntVarP(&limit, "limit", "l", 50, "Limit for pagination")
	orgsCmd.AddCommand(selectOrgCmd)

	deselectOrgCmd := &cobra.Command{
		Use:   "deselect",
		Short: "Deselect your current org",
		Long:  "Deselect your current org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.Deselect(cmd.Context())
		}),
	}
	orgsCmd.AddCommand(deselectOrgCmd)

	orgsCmd.AddCommand(&cobra.Command{
		Use:   "print-config",
		Short: "Print the current cli config",
		Long:  "Print the current cli config being used",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.PrintConfig(PrintJSON)
		}),
	})

	createInviteCmd := &cobra.Command{
		Use:   "invite",
		Short: "Invite a user to org",
		Long:  "Invite a user by email to org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.CreateInvite(cmd.Context(), email, PrintJSON)
		}),
	}
	createInviteCmd.Flags().StringVarP(&email, "email", "e", "", "Email of user to invite")
	orgsCmd.AddCommand(createInviteCmd)

	listInvitesCmd := &cobra.Command{
		Use:   "list-invites",
		Short: "List all org invites",
		Long:  "List all org invites",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.ListInvites(cmd.Context(), offset, limit, PrintJSON)
		}),
	}
	listInvitesCmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	listInvitesCmd.Flags().IntVarP(&limit, "limit", "l", 20, "Maximum invites to return")
	orgsCmd.AddCommand(listInvitesCmd)

	orgsCmd.AddCommand(c.orgWebhooksCmd())

	return orgsCmd
}

// interestsJSONHelp documents the shape accepted by --interests-json /
// --interests-file. Mirrors internal/pkg/interests/types.go (SubOps map and
// Outcome constants) and the "Filtering events with interests" section of
// docs/guides/webhooks.mdx; update all three in lockstep.
const interestsJSONHelp = `INTERESTS JSON SHAPE

  Top-level keys (one of):
    all_events: bool             subscribe to every supported event
    resources:  map<kind, cfg>   per-resource opt-in (omit kinds you don't want)

  Resource kinds and the ops they support:
    installs                provision, deprovision, reprovision
    components              deploy, teardown, drift
    sandboxes               provision, reprovision, deprovision, drift
    install_configurations  inputs, secrets
    runners                 provision, reprovision, inactive
    actions                 run

  Per-resource cfg fields (all optional):
    ops                 list of sub-ops; empty/omitted = every sub-op
    outcome             "all" (default) | "completion" | "failures"
    approval_requests   bool, deliver workflow-step approval requests
                        (independent of outcome)
    approval_responses  bool, deliver workflow-step approval responses
                        (independent of outcome)
    drift_detected      bool, deliver a notification ONLY when drift is
                        actually detected during a drift scan (independent
                        of outcome). Only meaningful for resources with a
                        "drift" sub-op (components, sandboxes).

EXAMPLES

  Subscribe to every supported event (the default if --interests-json/-file is
  omitted):
    {"all_events": true}

  Per-resource opt-in matching the dashboard's "opted-out-of-AllEvents"
  baseline:
    {
      "resources": {
        "installs":               {"outcome": "completion", "approval_requests": true, "approval_responses": true},
        "components":             {"outcome": "completion", "approval_requests": true, "approval_responses": true, "drift_detected": true},
        "sandboxes":              {"outcome": "completion", "approval_requests": true, "approval_responses": true, "drift_detected": true},
        "install_configurations": {"outcome": "completion", "approval_requests": true, "approval_responses": true}
      }
    }

  Narrowly scoped — only component deploy failures:
    {
      "resources": {
        "components": {"ops": ["deploy"], "outcome": "failures"}
      }
    }
`

func (c *cli) orgWebhooksCmd() *cobra.Command {
	var (
		webhookURL    string
		webhookSecret string
		webhookID     string

		createInterests orgs.InterestsFlags
		updateInterests orgs.InterestsFlags
	)

	webhooksCmd := &cobra.Command{
		Use:               "webhooks",
		Short:             "Manage operation lifecycle webhooks for the current org",
		Long:              "Manage operation lifecycle webhooks delivered to your endpoints for the current org",
		PersistentPreRunE: c.persistentPreRunE,
	}

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List webhooks for the current org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.ListWebhooks(cmd.Context(), PrintJSON)
		}),
	}
	webhooksCmd.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a webhook for the current org",
		Long: `Create a webhook subscription for operation lifecycle events on the current org.

By default the webhook subscribes to every supported event. To filter, pass
either --interests-json '<json>' or --interests-file path/to/interests.json.

` + interestsJSONHelp,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			createInterests.AllEventsIsSet = cmd.Flags().Changed("all-events")
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.CreateWebhook(cmd.Context(), webhookURL, webhookSecret, createInterests, PrintJSON)
		}),
	}
	createCmd.Flags().StringVarP(&webhookURL, "url", "u", "", "Webhook URL (must be an absolute http/https URL)")
	createCmd.MarkFlagRequired("url")
	createCmd.Flags().StringVarP(&webhookSecret, "secret", "s", "", "Optional shared secret used to sign webhook payloads (HMAC-SHA256, X-Nuon-Signature header)")
	createCmd.Flags().BoolVar(&createInterests.AllEvents, "all-events", true, "Subscribe to every supported event (default). Mutually exclusive with --interests-json/--interests-file")
	createCmd.Flags().StringVar(&createInterests.InterestsJSON, "interests-json", "", "Inline JSON interests config (mutually exclusive with --all-events / --interests-file)")
	createCmd.Flags().StringVar(&createInterests.InterestsFile, "interests-file", "", "Path to a JSON file containing the interests config (mutually exclusive with --all-events / --interests-json)")
	webhooksCmd.AddCommand(createCmd)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update a webhook for the current org",
		Long: `Replace the interests filter on a webhook and optionally rotate its signing secret.

The webhook URL is immutable — delete and recreate to rename. The signing
secret is preserved unless --secret <new> is passed, in which case it is
rotated to the provided value.

` + interestsJSONHelp,
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			updateInterests.AllEventsIsSet = cmd.Flags().Changed("all-events")
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.UpdateWebhook(cmd.Context(), webhookID, webhookSecret, updateInterests, PrintJSON)
		}),
	}
	updateCmd.Flags().StringVar(&webhookID, "webhook-id", "", "The ID of the webhook to update")
	updateCmd.MarkFlagRequired("webhook-id")
	updateCmd.Flags().StringVarP(&webhookSecret, "secret", "s", "", "New signing secret. Omit to keep the existing secret unchanged")
	updateCmd.Flags().BoolVar(&updateInterests.AllEvents, "all-events", true, "Subscribe to every supported event. Mutually exclusive with --interests-json/--interests-file")
	updateCmd.Flags().StringVar(&updateInterests.InterestsJSON, "interests-json", "", "Inline JSON interests config (mutually exclusive with --all-events / --interests-file)")
	updateCmd.Flags().StringVar(&updateInterests.InterestsFile, "interests-file", "", "Path to a JSON file containing the interests config (mutually exclusive with --all-events / --interests-json)")
	webhooksCmd.AddCommand(updateCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a webhook for the current org",
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.DeleteWebhook(cmd.Context(), webhookID, PrintJSON)
		}),
	}
	deleteCmd.Flags().StringVar(&webhookID, "webhook-id", "", "The ID of the webhook to delete")
	deleteCmd.MarkFlagRequired("webhook-id")
	webhooksCmd.AddCommand(deleteCmd)

	return webhooksCmd
}
