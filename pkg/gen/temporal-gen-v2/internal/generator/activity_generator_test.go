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
	assert.Contains(t, code, "options := workflow.GetActivityOptions(ctx)")
	assert.Contains(t, code, "options.ScheduleToCloseTimeout = time.Duration(3600000000000)") // 1h in ns
	assert.Contains(t, code, "options.StartToCloseTimeout = time.Duration(1800000000000)")    // 30m in ns
	assert.Contains(t, code, "MaximumAttempts: int32(3)")

	// Check overrides
	assert.Contains(t, code, "if opt.TaskQueue != \"\" {")
	assert.Contains(t, code, "options.TaskQueue = opt.TaskQueue")
}

func TestGenerateActivityWithWrapperPrefix(t *testing.T) {
	tests := []struct {
		name                 string
		activityData         ActivityData
		expectedWrapperFunc  string
		expectedAwaitFunc    string
		notExpectedAwaitFunc string
	}{
		{
			name: "Activity with wrapper-prefix",
			activityData: ActivityData{
				Name:         "CreateQueueSignal",
				OriginalName: "createQueueSignal",
				Receiver:     "*Activities",
				Options: &parser.ActivityOptions{
					GenerateWrapper:     true,
					WrapperPrefix:       "QueueInternal",
					StartToCloseTimeout: time.Minute,
				},
				Params: []Param{
					{Name: "queueID", ExportedName: "QueueID", Type: "string"},
					{Name: "signal", ExportedName: "Signal", Type: "*Signal"},
				},
				OutputType: "*app.QueueSignal",
			},
			expectedWrapperFunc:  "func (a *Activities) QueueInternalCreateQueueSignal(ctx context.Context, req CreateQueueSignalRequest) (*app.QueueSignal, error)",
			expectedAwaitFunc:    "func AwaitCreateQueueSignal(ctx workflow.Context, input CreateQueueSignalRequest, opts ...*workflow.ActivityOptions) (*app.QueueSignal, error)",
			notExpectedAwaitFunc: "func AwaitQueueInternalCreateQueueSignal", // Should NOT have prefix
		},
		{
			name: "Activity without wrapper-prefix",
			activityData: ActivityData{
				Name:         "ProcessItem",
				OriginalName: "processItem",
				Receiver:     "*Service",
				Options: &parser.ActivityOptions{
					GenerateWrapper:     true,
					StartToCloseTimeout: time.Minute,
				},
				Params: []Param{
					{Name: "itemID", ExportedName: "ItemID", Type: "string"},
				},
				OutputType: "", // No return value other than error
			},
			expectedWrapperFunc:  "func (a *Service) ProcessItem(ctx context.Context, req ProcessItemRequest) error",
			expectedAwaitFunc:    "func AwaitProcessItem(ctx workflow.Context, input ProcessItemRequest, opts ...*workflow.ActivityOptions) error",
			notExpectedAwaitFunc: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := GenerateActivity(tt.activityData)
			require.NoError(t, err)

			code := string(output)

			// Verify wrapper function has the prefix
			assert.Contains(t, code, tt.expectedWrapperFunc, "Wrapper function should match expected signature")

			// Verify Await function does NOT have the prefix
			assert.Contains(t, code, tt.expectedAwaitFunc, "Await function should NOT include wrapper prefix")

			// Verify the prefixed Await function does NOT exist
			if tt.notExpectedAwaitFunc != "" {
				assert.NotContains(t, code, tt.notExpectedAwaitFunc, "Await function should NOT have prefix in name")
			}

			// Verify ExecuteActivity call uses the correct function name (with prefix if present)
			if tt.activityData.Options.GenerateWrapper && tt.activityData.Options.WrapperPrefix != "" {
				expectedExecCall := "(*Activities)." + tt.activityData.Options.WrapperPrefix + tt.activityData.Name
				assert.Contains(t, code, expectedExecCall, "ExecuteActivity should call the prefixed wrapper function")
			}
		})
	}
}
