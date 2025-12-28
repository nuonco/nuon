package generator

import (
	"testing"
	"time"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	// Check Exec
	assert.Contains(t, code, "func ExecMyWorkflow(ctx workflow.Context, input MyInput, opts ...*workflow.ChildWorkflowOptions) (workflow.ChildWorkflowFuture, error)")

	// Check wrapper signature with variadic opts
	assert.Contains(t, code, "func AwaitMyWorkflow(ctx workflow.Context, input MyInput, opts ...*workflow.ChildWorkflowOptions) (MyOutput, error)")

	// Check default options
	assert.Contains(t, code, "TaskQueue:                \"my-queue\"")
	assert.Contains(t, code, "WorkflowExecutionTimeout: time.Duration(86400000000000)") // 24h
	assert.Contains(t, code, "WorkflowTaskTimeout:      time.Duration(600000000000)")   // 10m
	assert.Contains(t, code, "WaitForCancellation:      true")

	// Check overrides logic
	assert.Contains(t, code, "if opt.TaskQueue != \"\" {")
	assert.Contains(t, code, "cwo.TaskQueue = opt.TaskQueue")

	// Check cron wrapper
	assert.Contains(t, code, "func AwaitMyWorkflowWithCron(ctx workflow.Context, input MyInput, cronSchedule string, opts ...*workflow.ChildWorkflowOptions) (MyOutput, error)")
	assert.Contains(t, code, "CronSchedule: cronSchedule")
}
