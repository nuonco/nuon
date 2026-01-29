package parser

import (
	"testing"
	"time"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tests := map[string]struct {
		comments []string
		expected *Annotation
		wantErr  bool
	}{
		"Basic Workflow": {
			comments: []string{
				"// @" + config.AnnotationPrefix + " workflow",
			},
			expected: &Annotation{
				Type:         "workflow",
				WorkflowOpts: &WorkflowOptions{},
			},
		},
		"Basic Activity": {
			comments: []string{
				"// @" + config.AnnotationPrefix + " activity",
			},
			expected: &Annotation{
				Type:         "activity",
				ActivityOpts: &ActivityOptions{},
			},
		},
		"Activity with timeouts": {
			comments: []string{
				"// @" + config.AnnotationPrefix + " activity",
				"// @schedule-to-close-timeout 1h",
				"// @start-to-close-timeout 30m",
			},
			expected: &Annotation{
				Type: "activity",
				ActivityOpts: &ActivityOptions{
					ScheduleToCloseTimeout: time.Hour,
					StartToCloseTimeout:    30 * time.Minute,
				},
			},
		},
		"Activity with max retries": {
			comments: []string{
				"// @" + config.AnnotationPrefix + " activity",
				"// @max-retries 5",
			},
			expected: &Annotation{
				Type: "activity",
				ActivityOpts: &ActivityOptions{
					MaxRetries:  5,
					RetryPolicy: true,
				},
			},
		},
		"Activity with all options": {
			comments: []string{
				"// @" + config.AnnotationPrefix + " activity",
				"// @schedule-to-close-timeout 10m",
				"// @start-to-close-timeout 5m",
				"// @max-retries 3",
			},
			expected: &Annotation{
				Type: "activity",
				ActivityOpts: &ActivityOptions{
					ScheduleToCloseTimeout: 10 * time.Minute,
					StartToCloseTimeout:    5 * time.Minute,
					MaxRetries:             3,
					RetryPolicy:            true,
				},
			},
		},
		"Invalid duration": {
			comments: []string{
				"// @" + config.AnnotationPrefix + " activity",
				"// @schedule-to-close-timeout invalid",
			},
			wantErr: true,
		},
		"Activity with by-field": {
			comments: []string{
				"// @" + config.AnnotationPrefix + " activity",
				"// @by-field ID",
			},
			expected: &Annotation{
				Type: "activity",
				ActivityOpts: &ActivityOptions{
					ByField: "ID",
				},
			},
		},
		"Activity with by-field-only": {
			comments: []string{
				"// @" + config.AnnotationPrefix + " activity",
				"// @by-field ID",
				"// @by-field-only",
			},
			expected: &Annotation{
				Type: "activity",
				ActivityOpts: &ActivityOptions{
					ByField:     "ID",
					ByFieldOnly: true,
				},
			},
		},
		"Invalid by-field-only without by-field": {
			comments: []string{
				"// @" + config.AnnotationPrefix + " activity",
				"// @by-field-only",
			},
			wantErr: true,
		},
		"Workflow with all options": {
			comments: []string{
				"// @" + config.AnnotationPrefix + " workflow",
				"// @execution-timeout 1h",
				"// @task-timeout 30m",
				"// @id-template workflow-{{.ID}}",
				"// @wait-for-cancellation true",
				"// @workflow-task-queue default",
				"// @options-callback GetOptions",
			},
			expected: &Annotation{
				Type: "workflow",
				WorkflowOpts: &WorkflowOptions{
					ExecutionTimeout:    time.Hour,
					TaskTimeout:         30 * time.Minute,
					IDTemplate:          "workflow-{{.ID}}",
					WaitForCancellation: true,
					TaskQueue:           "default",
					OptionsCallback:     "GetOptions",
				},
			},
		},
		"Workflow with id-generator": {
			comments: []string{
				"// @" + config.AnnotationPrefix + " workflow",
				"// @id-generator GetID",
			},
			expected: &Annotation{
				Type: "workflow",
				WorkflowOpts: &WorkflowOptions{
					IDGenerator: "GetID",
				},
			},
		},
		"Basic Query": {
			comments: []string{
				"// @" + config.AnnotationPrefix + " query",
			},
			expected: &Annotation{
				Type:      "query",
				QueryOpts: &QueryOptions{},
			},
		},
		"Basic Update": {
			comments: []string{
				"// @" + config.AnnotationPrefix + " update",
			},
			expected: &Annotation{
				Type:       "update",
				UpdateOpts: &UpdateOptions{},
			},
		},
		"Update with ID": {
			comments: []string{
				"// @" + config.AnnotationPrefix + " update",
				"// @id my-update-id",
			},
			expected: &Annotation{
				Type: "update",
				UpdateOpts: &UpdateOptions{
					ID: "my-update-id",
				},
			},
		},
		"Unknown argument": {
			comments: []string{
				"// @" + config.AnnotationPrefix + " activity",
				"// @unknown-arg foo",
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := Parse(tt.comments)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
