package generator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/parser"
)

func TestGenerateWorkflow(t *testing.T) {
	data := WorkflowData{
		Name:         "MyWorkflow",
		OriginalName: "myPackage.MyWorkflow",
		InputType:    "MyInput",
		OutputType:   "MyOutput",
		Options: &parser.WorkflowOptions{
			ExecutionTimeout:    24 * time.Hour,
			TaskTimeout:         10 * time.Minute,
			TaskQueue:           "my-queue",
			WaitForCancellation: true,
		},
	}

	output, err := GenerateWorkflow(data)
	require.NoError(t, err)

	code := string(output)
	assert.Contains(t, code, "func ExecMyWorkflow(ctx workflow.Context, input MyInput, opts ...*workflow.ChildWorkflowOptions) (workflow.ChildWorkflowFuture, error)")

	assert.Contains(t, code, "func AwaitMyWorkflow(ctx workflow.Context, input MyInput, opts ...*workflow.ChildWorkflowOptions) (MyOutput, error)")

	assert.Contains(t, code, "TaskQueue:                \"my-queue\"")
	assert.Contains(t, code, "WorkflowExecutionTimeout: time.Duration(86400000000000)")
	assert.Contains(t, code, "WorkflowTaskTimeout:      time.Duration(600000000000)")
	assert.Contains(t, code, "WaitForCancellation:      true")

	assert.Contains(t, code, "if opt.TaskQueue != \"\" {")
	assert.Contains(t, code, "cwo.TaskQueue = opt.TaskQueue")

	assert.Contains(t, code, "func AwaitMyWorkflowWithCron(ctx workflow.Context, input MyInput, cronSchedule string, opts ...*workflow.ChildWorkflowOptions) (MyOutput, error)")
	assert.Contains(t, code, "CronSchedule: cronSchedule")
}

func TestGenerateWorkflowWithCallerIDTemplate(t *testing.T) {
	data := WorkflowData{
		Name:         "WorkflowWithCallerID",
		OriginalName: "WorkflowWithCallerID",
		InputType:    "TaskInput",
		OutputType:   "",
		Options: &parser.WorkflowOptions{
			IDTemplate: "{{.CallerID}}-subtask-{{.Req.TaskID}}",
		},
	}

	output, err := GenerateWorkflow(data)
	require.NoError(t, err)

	code := string(output)

	assert.Contains(t, code, "CallerID string")

	assert.Contains(t, code, "CallerID: info.WorkflowExecution.ID")

	assert.Contains(t, code, "{{.CallerID}}-subtask-{{.Req.TaskID}}")

	assert.Contains(t, code, "Req      TaskInput")

	assert.Contains(t, code, "Info     *workflow.Info")
}

func TestGenerateWorkflowWithMemo(t *testing.T) {
	data := WorkflowData{
		Name:         "QueueWorkflow",
		OriginalName: "QueueWorkflow",
		InputType:    "QueueInput",
		OutputType:   "",
		Options: &parser.WorkflowOptions{
			Memo: map[string]string{
				"type": "queue",
			},
		},
	}

	output, err := GenerateWorkflow(data)
	require.NoError(t, err)

	code := string(output)

	assert.Contains(t, code, `Memo: map[string]interface{}{`)
	assert.Contains(t, code, `"type": "queue"`)

	assert.Contains(t, code, "for k, v := range opt.Memo")
	assert.Contains(t, code, "cwo.Memo[k] = v")
}

func TestGenerateWorkflowWithMultipleMemo(t *testing.T) {
	data := WorkflowData{
		Name:         "HandlerWorkflow",
		OriginalName: "HandlerWorkflow",
		InputType:    "HandlerInput",
		OutputType:   "",
		Options: &parser.WorkflowOptions{
			Memo: map[string]string{
				"type":  "queue-handler",
				"owner": "system",
			},
		},
	}

	output, err := GenerateWorkflow(data)
	require.NoError(t, err)

	code := string(output)

	assert.Contains(t, code, `"queue-handler"`)
	assert.Contains(t, code, `"owner"`)
	assert.Contains(t, code, `"system"`)
}

func TestGenerateWorkflowWithoutMemo(t *testing.T) {
	data := WorkflowData{
		Name:         "SimpleWorkflow",
		OriginalName: "SimpleWorkflow",
		InputType:    "SimpleInput",
		OutputType:   "",
		Options:      &parser.WorkflowOptions{},
	}

	output, err := GenerateWorkflow(data)
	require.NoError(t, err)

	code := string(output)

	assert.NotContains(t, code, `Memo: map[string]interface{}{`)

	assert.Contains(t, code, "for k, v := range opt.Memo")
}
