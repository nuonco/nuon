package slackrender

import (
	"strings"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

func blocksText(msg Message) string {
	var sb strings.Builder
	for _, b := range msg.Blocks {
		switch blk := b.(type) {
		case *slack.HeaderBlock:
			if blk.Text != nil {
				sb.WriteString(blk.Text.Text + "\n")
			}
		case *slack.SectionBlock:
			if blk.Text != nil {
				sb.WriteString(blk.Text.Text + "\n")
			}
			for _, f := range blk.Fields {
				sb.WriteString(f.Text + "\n")
			}
		case *slack.ContextBlock:
			for _, el := range blk.ContextElements.Elements {
				if tb, ok := el.(*slack.TextBlockObject); ok {
					sb.WriteString(tb.Text + "\n")
				}
			}
		}
	}
	return sb.String()
}

func TestParentByFieldMention(t *testing.T) {
	e := Event{
		Kind:       KindWorkflow,
		Transition: TransitionStarted,
		OrgName:    "Acme",
		Workflow: WorkflowRef{
			ID:             "wfl_1",
			Type:           WorkflowTypeProvision,
			CreatedByEmail: "a@b.co",
		},
	}

	plain := blocksText(BuildParentMessage(e, time.Time{}))
	assert.Contains(t, plain, "a@b.co")
	assert.NotContains(t, plain, "<@")

	e.SlackUserIDByEmail = map[string]string{"a@b.co": "U123"}
	mentioned := blocksText(BuildParentMessage(e, time.Time{}))
	assert.Contains(t, mentioned, "<@U123>")
	assert.NotContains(t, mentioned, "a@b.co")
}

func TestApprovalResponderMention(t *testing.T) {
	e := Event{
		Kind:       KindWorkflowStepApproval,
		Transition: TransitionApproved,
		Workflow:   WorkflowRef{ID: "wfl_1", Type: WorkflowTypeProvision},
		Step:       &StepRef{ID: "stp_1", Name: "terraform-apply"},
		Approval:   &ApprovalRef{ID: "apr_1", RespondedBy: "a@b.co"},
	}

	assert.Contains(t, blocksText(BuildParentRollup(e, time.Time{})), "by a@b.co")
	assert.Contains(t, blocksText(BuildChildMessage(e)), "by a@b.co")

	e.SlackUserIDByEmail = map[string]string{"a@b.co": "U123"}
	assert.Contains(t, blocksText(BuildParentRollup(e, time.Time{})), "by <@U123>")
	assert.Contains(t, blocksText(BuildChildMessage(e)), "by <@U123>")
}

func TestAccountIDFallbackRendersLiterally(t *testing.T) {
	e := Event{
		Kind:       KindWorkflowStepApproval,
		Transition: TransitionRejected,
		Workflow:   WorkflowRef{ID: "wfl_1", Type: WorkflowTypeProvision},
		Step:       &StepRef{ID: "stp_1", Name: "terraform-apply"},
		Approval:   &ApprovalRef{ID: "apr_1", RespondedBy: "acct123"},
	}
	out := blocksText(BuildChildMessage(e))
	assert.Contains(t, out, "by acct123")
	assert.NotContains(t, out, "<@")
}

func TestChildContextEscapesOnce(t *testing.T) {
	e := Event{
		Kind:       KindWorkflowStep,
		Transition: TransitionFailed,
		Workflow:   WorkflowRef{ID: "wfl_1", Type: WorkflowTypeProvision},
		Step:       &StepRef{ID: "stp_1", Name: "apply"},
		Outcome:    &Outcome{Status: "failed", Error: "x < y"},
	}
	out := blocksText(BuildChildMessage(e))
	assert.Contains(t, out, "x &lt; y")
	assert.NotContains(t, out, "&amp;lt;")
}

func TestFlatMessageEscapesOnce(t *testing.T) {
	e := Event{
		Kind:       KindWorkflow,
		Transition: TransitionStarted,
		OrgName:    "A&B",
		Workflow:   WorkflowRef{ID: "wfl_1", Type: WorkflowTypeProvision},
	}
	out := blocksText(BuildFlatMessage(e))
	assert.Contains(t, out, "A&amp;B")
	assert.NotContains(t, out, "&amp;amp;")
}
