package generator

import (
	"testing"
	"time"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateActivity(t *testing.T) {
	data := ActivityData{
		Name:         "MyActivity",
		OriginalName: "myPackage.MyActivity",
		InputType:    "MyInput",
		OutputType:   "MyOutput",
		Options: &parser.ActivityOptions{
			ScheduleToCloseTimeout: time.Hour,
			StartToCloseTimeout:    30 * time.Minute,
			MaxRetries:             3,
		},
	}

	output, err := GenerateActivity(data)
	require.NoError(t, err)

	code := string(output)
	assert.Contains(t, code, "func AwaitMyActivity(ctx workflow.Context, input MyInput, opts ...*workflow.ActivityOptions) (MyOutput, error)")
	assert.Contains(t, code, "ScheduleToCloseTimeout: time.Duration(3600000000000)") // 1h in ns
	assert.Contains(t, code, "StartToCloseTimeout:    time.Duration(1800000000000)") // 30m in ns
	assert.Contains(t, code, "MaximumAttempts: int32(3)")

	// Check overrides
	assert.Contains(t, code, "if opt.TaskQueue != \"\" {")
	assert.Contains(t, code, "options.TaskQueue = opt.TaskQueue")
}
