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

// subscriptionJSONHelp documents the shape accepted by --subscription-json
// and --subscription-file. The unified payload combines the events filter
// (interests) and the optional scope predicate (match) into a single
// committable artifact. Mirrors:
//   - services/ctl-api/internal/pkg/interests/types.go (SubOps map + Outcome constants)
//   - pkg/labels/match.go (SubscriptionMatch / TargetMatch / Selector)
//   - the "Filtering events with interests" + "Scoping deliveries" sections
//     of docs/guides/webhooks.mdx
//
// Update all four in lockstep.
const subscriptionJSONHelp = `SUBSCRIPTION JSON SHAPE

  Top-level keys (both optional):
    interests  events filter — which events fire for this webhook
    match      scope predicate — which entities are in scope (omit for
               org-wide / every entity)

  Omitting --subscription-json / --subscription-file entirely is equivalent
  to {"interests": {"all_events": true}} — every event in the org, no
  scoping. On update both fields are replaced wholesale; pass the existing
  shape alongside any edits to preserve it.

INTERESTS

  Top-level keys (one of):
    all_events: bool             subscribe to every supported event
    resources:  map<kind, cfg>   per-resource opt-in (omit kinds you don't want)

  Resource kinds and the ops they support:
    installs                provision, deprovision, reprovision
    components              deploy, teardown
    sandboxes               provision, reprovision, deprovision
    install_configurations  inputs, secrets
    runners                 provision, reprovision, inactive
    actions                 run

  Note: to subscribe to drift, use "drift_detected" (below) — it fires only when
  drift is actually found, not on every clean scan.

  Per-resource cfg fields (all optional):
    ops                 list of sub-ops; empty/omitted = every sub-op
    outcome             "all" (default) | "completion" | "failures" | "none"
                        ("none" mutes lifecycle for this resource — combine
                        with drift_detected / approval_requests /
                        approval_responses for drift-only or approvals-only
                        subscriptions)
    approval_requests   bool, deliver workflow-step approval requests
                        (independent of outcome)
    approval_responses  bool, deliver workflow-step approval responses
                        (independent of outcome)
    drift_detected      bool, deliver a notification ONLY when drift is
                        actually detected during a drift scan (independent
                        of outcome). Only meaningful for components and
                        sandboxes.

MATCH

  Top-level keys (any combination; OR across kinds):
    installs    target match for install-scoped events
    components  target match for component-scoped events
    actions     target match for action-scoped events

  Per-kind TargetMatch fields (all optional; an empty {} means "any entity
  of this kind"):
    ids       list of entity IDs (OR within the list)
    selector  label selector with match_labels: {key: value} (AND across keys;
              value "*" means "key must exist")

  Composition: the match matches an event when ANY populated kind matches.
  Within a kind, the entity matches when its ID is in ids OR the selector
  hits its labels.

EXAMPLES

  Subscribe to every supported event in the org (equivalent to omitting the
  flag entirely):
    {"interests": {"all_events": true}}

  Component deploy failures across the whole org:
    {
      "interests": {
        "resources": {
          "components": {"ops": ["deploy"], "outcome": "failures"}
        }
      }
    }

  Every event, but only for two specific installs:
    {
      "interests": {"all_events": true},
      "match":     {"installs": {"ids": ["ins_abc123", "ins_def456"]}}
    }

  Drift-only on components carrying env=prod:
    {
      "interests": {
        "resources": {
          "components": {"outcome": "none", "drift_detected": true}
        }
      },
      "match": {
        "components": {"selector": {"match_labels": {"env": "prod"}}}
      }
    }

  Per-resource baseline — terminal events plus approvals for the four most
  common resources, scoped to a label-selected install:
    {
      "interests": {
        "resources": {
          "installs":               {"outcome": "completion", "approval_requests": true, "approval_responses": true},
          "components":             {"outcome": "completion", "approval_requests": true, "approval_responses": true, "drift_detected": true},
          "sandboxes":              {"outcome": "completion", "approval_requests": true, "approval_responses": true, "drift_detected": true},
          "install_configurations": {"outcome": "completion", "approval_requests": true, "approval_responses": true}
        }
      },
      "match": {
        "installs": {"selector": {"match_labels": {"team": "platform"}}}
      }
    }
`

func (c *cli) orgWebhooksCmd() *cobra.Command {
	var (
		webhookURL    string
		webhookSecret string
		webhookID     string

		createSubscription orgs.SubscriptionFlags
		updateSubscription orgs.SubscriptionFlags
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

When neither --subscription-json nor --subscription-file is passed in an
interactive terminal, the CLI drops into a picker that walks you through
the same fields the dashboard's Slack subscription modal collects.

In non-interactive sessions (NUON_NO_TTY=true, CI, piped stdout) the
default is to subscribe to every supported event in the org. To narrow
events and/or scope the webhook to specific entities non-interactively,
pass either --subscription-json '<json>' or --subscription-file
path/to/subscription.json.

` + subscriptionJSONHelp,
		Annotations: tuiAnnotation(TUIContextual),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.CreateWebhook(cmd.Context(), webhookURL, webhookSecret, createSubscription, PrintJSON)
		}),
	}
	createCmd.Flags().StringVarP(&webhookURL, "url", "u", "", "Webhook URL (must be an absolute http/https URL)")
	createCmd.MarkFlagRequired("url")
	createCmd.Flags().StringVarP(&webhookSecret, "secret", "s", "", "Optional shared secret used to sign webhook payloads (HMAC-SHA256, X-Nuon-Signature header)")
	createCmd.Flags().StringVar(&createSubscription.JSON, "subscription-json", "", "Inline JSON subscription describing interests + match (mutually exclusive with --subscription-file)")
	createCmd.Flags().StringVar(&createSubscription.File, "subscription-file", "", "Path to a JSON file containing the subscription describing interests + match (mutually exclusive with --subscription-json)")
	webhooksCmd.AddCommand(createCmd)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update a webhook for the current org",
		Long: `Replace the subscription on a webhook and optionally rotate its signing secret.

The webhook URL is immutable — delete and recreate to rename. The signing
secret is preserved unless --secret <new> is passed, in which case it is
rotated to the provided value.

The subscription (interests + match) is replaced wholesale on every update.

When neither --subscription-json nor --subscription-file is passed in an
interactive terminal, the CLI drops into a picker so you can rebuild the
subscription from scratch (the previous subscription is NOT pre-loaded —
the picker starts at "all events" / "every entity").

In non-interactive sessions (NUON_NO_TTY=true, CI, piped stdout) omitting
both flags resets the webhook to the implicit default (every event in
the org, no scoping). Pass the existing subscription alongside any edits
to preserve it.

` + subscriptionJSONHelp,
		Annotations: tuiAnnotation(TUIContextual),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := orgs.New(c.apiClient, c.cfg)
			return svc.UpdateWebhook(cmd.Context(), webhookID, webhookSecret, updateSubscription, PrintJSON)
		}),
	}
	updateCmd.Flags().StringVar(&webhookID, "webhook-id", "", "The ID of the webhook to update")
	updateCmd.MarkFlagRequired("webhook-id")
	updateCmd.Flags().StringVarP(&webhookSecret, "secret", "s", "", "New signing secret. Omit to keep the existing secret unchanged")
	updateCmd.Flags().StringVar(&updateSubscription.JSON, "subscription-json", "", "Inline JSON subscription describing interests + match (mutually exclusive with --subscription-file)")
	updateCmd.Flags().StringVar(&updateSubscription.File, "subscription-file", "", "Path to a JSON file containing the subscription describing interests + match (mutually exclusive with --subscription-json)")
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
