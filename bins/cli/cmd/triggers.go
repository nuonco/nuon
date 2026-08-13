package cmd

import (
	"encoding/json"
	"time"

	"github.com/spf13/cobra"

	triggersservice "github.com/nuonco/nuon/bins/cli/internal/services/triggers"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *cli) triggerEventsCmd() *cobra.Command {
	var limit int
	var trigger string
	var cursor string
	var eventID string
	var last bool
	var appConfig string
	var pollInterval time.Duration
	var pathsLast bool
	var tailRaw bool
	var eventType, outcome, search, receivedAfter, receivedBefore string

	eventsCmd := &cobra.Command{Use: "events", Short: "Inspect and replay received events"}
	listCmd := &cobra.Command{
		Use: "list", Aliases: []string{"ls"}, Short: "List trigger events",
		Args:        cobra.NoArgs,
		Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.triggers.List(cmd.Context(), models.TriggerEventListQuery{Limit: limit, Trigger: trigger, EventType: eventType, Outcome: outcome, Search: search, ReceivedAfter: receivedAfter, ReceivedBefore: receivedBefore, Cursor: cursor}, PrintJSON)
		}),
	}
	listCmd.Flags().IntVarP(&limit, "limit", "l", 50, "Maximum events to return")
	listCmd.Flags().StringVar(&trigger, "trigger", "", "Trigger ID or name")
	listCmd.MarkFlagRequired("trigger")
	listCmd.Flags().StringVar(&cursor, "cursor", "", "Continue listing from an opaque cursor")
	listCmd.Flags().StringVar(&eventType, "event-type", "", "Filter by exact event type")
	listCmd.Flags().StringVar(&outcome, "outcome", "", "Filter by outcome: ok, ignored, rejected, processing, or failed")
	listCmd.Flags().StringVar(&search, "query", "", "Search by Nuon event ID or external provider ID")
	listCmd.Flags().StringVar(&receivedAfter, "received-after", "", "Filter events received at or after an RFC3339 timestamp")
	listCmd.Flags().StringVar(&receivedBefore, "received-before", "", "Filter events received before an RFC3339 timestamp")

	getCmd := &cobra.Command{
		Use: "get <event-id>", Short: "Get a trigger event",
		Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return ui.PrintError(&ui.CLIUserError{Msg: "get requires exactly one event ID"})
			}
			return c.triggers.Get(cmd.Context(), args[0], PrintJSON)
		}),
	}
	replayCmd := &cobra.Command{
		Use: "replay <event-id>", Short: "Replay a trigger event",
		Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return ui.PrintError(&ui.CLIUserError{Msg: "replay requires exactly one event ID"})
			}
			return c.triggers.Replay(cmd.Context(), args[0], PrintJSON)
		}),
	}
	pathsCmd := &cobra.Command{
		Use: "paths [event-id]", Short: "List filterable payload paths and request headers",
		Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent),
		Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 || (len(args) == 1) == pathsLast {
				return ui.PrintError(&ui.CLIUserError{Msg: "paths requires exactly one event ID or --last"})
			}
			if pathsLast && trigger == "" {
				return ui.PrintError(&ui.CLIUserError{Msg: "--last requires --trigger"})
			}
			id := ""
			if len(args) == 1 {
				id = args[0]
			}
			return c.triggers.Paths(cmd.Context(), id, pathsLast, trigger, PrintJSON)
		}),
	}
	pathsCmd.Flags().BoolVar(&pathsLast, "last", false, "Inspect the most recent event")
	pathsCmd.Flags().StringVar(&trigger, "trigger", "", "Trigger ID or name (required with --last)")
	testCmd := &cobra.Command{
		Use: "test", Short: "Test a local trigger config against an event",
		Args:        cobra.NoArgs,
		Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			if (eventID == "") == !last {
				return ui.PrintError(triggersservice.SelectorError())
			}
			if last && trigger == "" {
				return ui.PrintError(&ui.CLIUserError{Msg: "--last requires --trigger"})
			}
			if appConfig == "" {
				return ui.PrintError(triggersservice.AppConfigError())
			}
			return c.triggers.Test(cmd.Context(), eventID, last, trigger, appConfig, PrintJSON)
		}),
	}
	testCmd.Flags().StringVar(&eventID, "event", "", "Event ID to test")
	testCmd.Flags().BoolVar(&last, "last", false, "Test the most recent event")
	testCmd.Flags().StringVar(&trigger, "trigger", "", "Trigger ID or name (required with --last)")
	testCmd.Flags().StringVar(&appConfig, "app-config", "", "Path to the local app TOML config")
	tailCmd := &cobra.Command{
		Use: "tail", Short: "Continuously print newly received trigger events", Args: cobra.NoArgs,
		Annotations: outputsAnnotation(OutputTable),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			return c.triggers.Tail(cmd.Context(), trigger, pollInterval, tailRaw)
		}),
	}
	tailCmd.Flags().StringVar(&trigger, "trigger", "", "Trigger ID or name")
	tailCmd.MarkFlagRequired("trigger")
	tailCmd.Flags().DurationVar(&pollInterval, "poll-interval", 2*time.Second, "Polling interval")
	tailCmd.Flags().BoolVar(&tailRaw, "raw", false, "Show raw request bodies")
	eventsCmd.AddCommand(listCmd, getCmd, replayCmd, pathsCmd, testCmd, tailCmd)
	return eventsCmd
}

func (c *cli) triggerDispatchesCmd() *cobra.Command {
	var limit int
	var eventID string
	var cursor string

	dispatchesCmd := &cobra.Command{Use: "dispatches", Short: "Inspect and retry trigger dispatches"}
	dispatchListCmd := &cobra.Command{Use: "list", Aliases: []string{"ls"}, Short: "List trigger dispatches", Args: cobra.NoArgs, Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent), Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
		return c.triggers.ListDispatches(cmd.Context(), limit, eventID, cursor, PrintJSON)
	})}
	dispatchListCmd.Flags().IntVarP(&limit, "limit", "l", 50, "Maximum dispatches to return")
	dispatchListCmd.Flags().StringVar(&eventID, "event-id", "", "Filter by event ID")
	dispatchListCmd.Flags().StringVar(&cursor, "cursor", "", "Continue listing from an opaque cursor")
	dispatchesCmd.AddCommand(dispatchListCmd, &cobra.Command{Use: "get <dispatch-id>", Short: "Get a trigger dispatch", Args: cobra.ExactArgs(1), Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent), Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
		return c.triggers.GetDispatch(cmd.Context(), args[0], PrintJSON)
	})}, &cobra.Command{Use: "retry <dispatch-id>", Short: "Retry a trigger dispatch", Args: cobra.ExactArgs(1), Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent), Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
		return c.triggers.RetryDispatch(cmd.Context(), args[0], PrintJSON)
	})})
	return dispatchesCmd
}

func (c *cli) triggersCmd() *cobra.Command {
	var description, preset, signingSecret, authType, envelope, typeHeader, typePayload, idHeader, idPayload string
	var audience []string
	var expectedEmail, expectedSubject, topicARN string
	var authConfig string
	var force bool

	triggersCmd := &cobra.Command{Use: "triggers", Short: "Manage event triggers", GroupID: AdditionalGroup.ID}
	createTriggerCmd := &cobra.Command{Use: "create <name>", Short: "Create an event trigger", Args: cobra.ExactArgs(1), Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent), Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
		req := &models.TriggerCreateRequest{Name: args[0], Description: description, Preset: preset, Secret: signingSecret, AuthType: authType, Envelope: envelope, TypeFrom: models.EventFieldSelector{Header: typeHeader, Payload: typePayload}, IDFrom: models.EventFieldSelector{Header: idHeader, Payload: idPayload}}
		if authConfig != "" {
			if err := json.Unmarshal([]byte(authConfig), &req.AuthConfig); err != nil {
				return ui.PrintError(&ui.CLIUserError{Msg: "--auth-config must be a JSON object: " + err.Error()})
			}
		}
		if len(audience) != 0 {
			req.AuthConfig.Audience = audience
		}
		if expectedEmail != "" {
			req.AuthConfig.ExpectedEmail = expectedEmail
		}
		if expectedSubject != "" {
			req.AuthConfig.ExpectedSubject = expectedSubject
		}
		if topicARN != "" {
			req.AuthConfig.TopicARN = topicARN
		}
		return c.triggers.CreateTrigger(cmd.Context(), req, PrintJSON)
	})}
	createTriggerCmd.Flags().StringVar(&description, "description", "", "Trigger description")
	createTriggerCmd.Flags().StringVar(&preset, "preset", "", "Provider preset: github, gitlab, bitbucket, gitea, forgejo, terraform-cloud, google-pubsub, azure-devops, aws-eventbridge, aws-sns, azure-event-grid, slack-events, or datadog")
	createTriggerCmd.Flags().StringVar(&signingSecret, "signing-secret", "", "Slack signing secret (slack-events preset only)")
	createTriggerCmd.Flags().StringVar(&authType, "auth-type", "", "Authentication type: none, hmac, api_key, basic, bearer_jwt, or sns_signature")
	createTriggerCmd.Flags().StringVar(&authConfig, "auth-config", "", "Authentication configuration as JSON")
	createTriggerCmd.Flags().StringSliceVar(&audience, "audience", nil, "OIDC audience (repeatable or comma-separated)")
	createTriggerCmd.Flags().StringVar(&expectedEmail, "expected-email", "", "Expected OIDC token email")
	createTriggerCmd.Flags().StringVar(&expectedSubject, "expected-subject", "", "Expected OIDC token subject")
	createTriggerCmd.Flags().StringVar(&topicARN, "topic-arn", "", "Allowed AWS SNS topic ARN")
	createTriggerCmd.Flags().StringVar(&envelope, "envelope", "", "Event envelope: none, pubsub_push, cloudevents, or sns")
	createTriggerCmd.Flags().StringVar(&typeHeader, "type-header", "", "Header containing the event type")
	createTriggerCmd.Flags().StringVar(&typePayload, "type-payload", "", "Payload path containing the event type")
	createTriggerCmd.Flags().StringVar(&idHeader, "id-header", "", "Header containing the event ID")
	createTriggerCmd.Flags().StringVar(&idPayload, "id-payload", "", "Payload path containing the event ID")
	triggerAction := func(use, short string, fn func(*triggersservice.Service, *cobra.Command, string) error) *cobra.Command {
		return &cobra.Command{Use: use + " <trigger-id>", Short: short, Args: cobra.ExactArgs(1), Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent), Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
			return fn(c.triggers, cmd, args[0])
		})}
	}
	listTriggers := &cobra.Command{Use: "list", Aliases: []string{"ls"}, Short: "List event triggers", Args: cobra.NoArgs, Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent), Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
		return c.triggers.ListTriggers(cmd.Context(), PrintJSON)
	})}
	getTrigger := triggerAction("get", "Get an event trigger", func(s *triggersservice.Service, cmd *cobra.Command, id string) error {
		return s.GetTrigger(cmd.Context(), id, PrintJSON)
	})
	rotate := triggerAction("rotate-secret", "Rotate a trigger secret", func(s *triggersservice.Service, cmd *cobra.Command, id string) error {
		return s.RotateTriggerSecret(cmd.Context(), id, PrintJSON)
	})
	enable := triggerAction("enable", "Enable an event trigger", func(s *triggersservice.Service, cmd *cobra.Command, id string) error {
		return s.SetTrigger(cmd.Context(), id, true, PrintJSON)
	})
	disable := triggerAction("disable", "Disable an event trigger", func(s *triggersservice.Service, cmd *cobra.Command, id string) error {
		return s.SetTrigger(cmd.Context(), id, false, PrintJSON)
	})
	revoke := &cobra.Command{Use: "revoke-secret <trigger-id> <secret-id>", Short: "Revoke a trigger secret", Args: cobra.ExactArgs(2), Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent), Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
		return c.triggers.RevokeTriggerSecret(cmd.Context(), args[0], args[1], PrintJSON)
	})}
	revealSecret := &cobra.Command{Use: "reveal-secret <trigger-id> <secret-id>", Short: "Reveal an active trigger secret", Args: cobra.ExactArgs(2), Annotations: outputsAnnotation(OutputTable, OutputJSON, OutputAgent), Run: c.wrapCmd(func(cmd *cobra.Command, args []string) error {
		return c.triggers.RevealTriggerSecret(cmd.Context(), args[0], args[1], PrintJSON)
	})}
	revealURL := triggerAction("reveal-ingress-url", "Reveal the current ingress URL", func(s *triggersservice.Service, cmd *cobra.Command, id string) error {
		return s.RevealTriggerIngressURL(cmd.Context(), id, PrintJSON)
	})
	replaceURL := triggerAction("replace-ingress-url", "Replace the ingress URL", func(s *triggersservice.Service, cmd *cobra.Command, id string) error {
		return s.ReplaceTriggerIngressURL(cmd.Context(), id, PrintJSON)
	})
	deleteTrigger := triggerAction("delete", "Delete an event trigger", func(s *triggersservice.Service, cmd *cobra.Command, id string) error {
		return s.DeleteTrigger(cmd.Context(), id, force, PrintJSON)
	})
	deleteTrigger.Flags().BoolVar(&force, "force", false, "Delete even when referenced by trigger rules")
	triggersCmd.AddCommand(createTriggerCmd, listTriggers, getTrigger, rotate, enable, disable, revoke, revealSecret, revealURL, replaceURL, deleteTrigger, c.triggerEventsCmd(), c.triggerDispatchesCmd())
	return triggersCmd
}
