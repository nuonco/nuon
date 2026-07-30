package activities

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestRenderRunbookInputDefersRuntimeOutputs(t *testing.T) {
	field := `{{.runbook_inputs.prefix}}-{{.runbook_outputs.wait_for_tag.event.payload.tag}}`

	rendered, err := renderRunbookInput(field, map[string]any{
		"runbook_inputs": map[string]string{"prefix": "release"},
	})

	require.NoError(t, err)
	require.Equal(t, field, rendered)
}

func TestRenderRunbookRuntimeFieldUsesMatchedEvent(t *testing.T) {
	data := map[string]any{
		"runbook_inputs": map[string]string{"prefix": "release"},
		"runbook_outputs": map[string]any{
			"wait_for_tag": map[string]any{
				"event": map[string]any{
					"payload": map[string]any{"tag": "v1.2.3"},
				},
			},
		},
	}

	rendered, err := renderRunbookRuntimeField(
		`{{.runbook_inputs.prefix}}={{.runbook_outputs.wait_for_tag.event.payload.tag}}`,
		data,
	)

	require.NoError(t, err)
	require.Equal(t, "release=v1.2.3", rendered)
}

func TestRenderRunbookHstoreSupportsStepNamesWithDashes(t *testing.T) {
	value := `{{(index .runbook_outputs "wait-for-tag").event.payload.tag}}`
	values := pgtype.Hstore{"GAR_TAG": &value}
	data := map[string]any{
		"runbook_outputs": map[string]any{
			"wait-for-tag": map[string]any{
				"event": map[string]any{
					"payload": map[string]any{"tag": "v1.2.3"},
				},
			},
		},
	}

	rendered, err := renderRunbookHstore(values, data)

	require.NoError(t, err)
	require.Equal(t, "v1.2.3", *rendered["GAR_TAG"])
}

func TestRunbookActionNeedsRuntimeRendering(t *testing.T) {
	rendered := "v1.2.3"
	require.False(t, runbookActionNeedsRuntimeRendering(&app.InstallActionWorkflowRun{
		RunEnvVars: pgtype.Hstore{"GAR_TAG": &rendered},
	}))

	template := `{{.runbook_outputs.wait.event.payload.tag}}`
	require.True(t, runbookActionNeedsRuntimeRendering(&app.InstallActionWorkflowRun{
		RunEnvVars: pgtype.Hstore{"GAR_TAG": &template},
	}))
}

func TestRenderRunbookStepRendersEventFilterInput(t *testing.T) {
	step := app.RunbookStepConfig{
		Filters: []app.TriggerFilter{{
			Path:  "$.tag",
			Op:    app.TriggerFilterTypeEq,
			Value: "{{.runbook_inputs.image_tag}}",
		}},
	}

	err := renderRunbookStep(&step, map[string]any{
		"runbook_inputs": map[string]string{"image_tag": "example.test/image:v2"},
	})

	require.NoError(t, err)
	require.Equal(t, "example.test/image:v2", step.Filters[0].Value)
}

func TestRenderRunbookStepRejectsMissingEventFilterInput(t *testing.T) {
	step := app.RunbookStepConfig{
		Filters: []app.TriggerFilter{{
			Path:  "$.tag",
			Op:    app.TriggerFilterTypeEq,
			Value: "{{.runbook_inputs.missing}}",
		}},
	}

	err := renderRunbookStep(&step, map[string]any{"runbook_inputs": map[string]string{}})

	require.Error(t, err)
}

func TestRunbookEventOutputUsesNormalizedEvent(t *testing.T) {
	occurredAt := time.Date(2026, time.July, 27, 12, 30, 0, 0, time.UTC)
	event := &app.TriggerEvent{
		ID:             "event-id",
		ExternalID:     "provider-id",
		Source:         "urn:producer",
		TriggerID:      "trigger-id",
		Trigger:        app.Trigger{Name: "gar"},
		EventType:      "tag.updated",
		OccurredAt:     &occurredAt,
		ReceivedAt:     occurredAt.Add(time.Second),
		Payload:        json.RawMessage(`{"tag":"v1.2.3","generation":9007199254740993}`),
		Headers:        map[string][]string{"Authorization": {"secret"}},
		RawBody:        []byte("provider envelope"),
		RawBodySHA256:  "raw-sha",
		PayloadSHA256:  "payload-sha",
		RawContentType: "application/json",
	}

	output, err := runbookEventOutput(event)
	require.NoError(t, err)
	payload := output["payload"].(map[string]any)
	require.Equal(t, "v1.2.3", payload["tag"])
	require.Equal(t, json.Number("9007199254740993"), payload["generation"])
	require.Equal(t, "urn:producer", output["source"])
	require.Equal(t, "tag.updated", output["type"])
	require.Equal(t, map[string]any{"id": "trigger-id", "name": "gar"}, output["trigger"])
	require.NotContains(t, output, "headers")
	require.NotContains(t, output, "raw_body")
}
