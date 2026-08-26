package parser

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/config"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/tags"
)

type ActivityOptions struct {
	Namespace              string
	TaskQueue              string
	ScheduleToCloseTimeout time.Duration
	ScheduleToStartTimeout time.Duration
	StartToCloseTimeout    time.Duration
	HeartbeatTimeout       time.Duration
	WaitForCancellation    bool
	ActivityID             string
	MaxRetries             int
	RetryPolicy            bool
	DisableEagerExecution  bool
	OptionsCallback        string
	ByField                string
	ByFieldOnly            bool
	GenerateWrapper        bool
	WrapperPrefix          string // Prefix to add to generated wrapper function name
	ReplicaRead            bool
	IsLocal                bool
}

type WorkflowOptions struct {
	ExecutionTimeout    time.Duration
	TaskTimeout         time.Duration
	IDGenerator         string
	IDTemplate          string
	WaitForCancellation bool
	TaskQueue           string
	OptionsCallback     string
	Memo                map[string]string
}

type QueryOptions struct{}

type SignalOptions struct{}

type UpdateOptions struct {
	ID string
}

type Annotation struct {
	Type         string
	Tags         []string
	ActivityOpts *ActivityOptions
	WorkflowOpts *WorkflowOptions
	QueryOpts    *QueryOptions
	SignalOpts   *SignalOptions
	UpdateOpts   *UpdateOptions
}

// Validate checks if the annotation configuration is valid
func (a *Annotation) Validate() error {
	if a.ActivityOpts != nil {
		if a.ActivityOpts.ByFieldOnly && a.ActivityOpts.ByField == "" {
			return fmt.Errorf("@by-field-only requires @by-field to be specified")
		}
		if a.ActivityOpts.WrapperPrefix != "" && !a.ActivityOpts.GenerateWrapper {
			return fmt.Errorf("@wrapper-prefix requires @as-wrapper to be specified")
		}
		if a.ActivityOpts.ReplicaRead && !a.ActivityOpts.GenerateWrapper {
			return fmt.Errorf("@replica-read requires @as-wrapper to be specified")
		}
	}
	if a.WorkflowOpts != nil {
		if a.WorkflowOpts.IDGenerator != "" && a.WorkflowOpts.IDTemplate != "" {
			return fmt.Errorf("@id-generator and @id-template may not be specified together")
		}
	}
	return nil
}

// TagError marks a failure to resolve the tags on a function: an unknown or
// duplicated tag, a `@tag` with no config to resolve it against, or a `@tag` on
// something other than an activity.
//
// It is distinguished from other parse errors because it is never recoverable:
// non-strict runs downgrade parse failures to a warning and skip the function,
// which for a mistyped tag would silently drop a wrapper the caller expects to
// exist. See file.ProcessFile.
type TagError struct {
	Err error
}

func (e *TagError) Error() string { return e.Err.Error() }
func (e *TagError) Unwrap() error { return e.Err }

func tagErrorf(format string, args ...any) error {
	return &TagError{Err: fmt.Errorf(format, args...)}
}

// Parse checks if a comment group contains the generator annotation.
//
// It applies no tag defaults; use ParseWithTags when a tag config is in play.
func Parse(comments []string) (*Annotation, error) {
	return ParseWithTags(comments, nil)
}

// ParseWithTags parses a comment group and folds in the defaults implied by
// any `@tag <name>` set on the function.
//
// Tags apply to activities only. Tag attributes are lowered into synthetic
// annotation lines that are parsed *ahead* of the function's own comments.
// Since parseLines is last-write-wins, that yields the precedence chain
// defaults -> tags (in source order) -> explicit annotations without a
// second assignment code path.
func ParseWithTags(comments []string, cfg *tags.Config) (*Annotation, error) {
	annotation, err := parseLines(comments)
	if err != nil || annotation == nil {
		return nil, err
	}

	if len(annotation.Tags) > 0 && annotation.Type != "activity" {
		return nil, tagErrorf("@tag is only supported on activities, found on %s", annotation.Type)
	}

	if annotation.Type == "activity" && (len(annotation.Tags) > 0 || cfg != nil) {
		names := annotation.Tags

		lines, err := cfg.AnnotationLines(names)
		if err != nil {
			return nil, &TagError{Err: err}
		}

		if len(lines) > 0 {
			merged := make([]string, 0, len(lines)+len(comments)+1)
			// Re-state the marker so the synthetic lines land inside an
			// annotated block. The duplicate marker in `comments` is a no-op:
			// parseLines only honours the first one.
			merged = append(merged, "// @"+config.AnnotationPrefix+" "+annotation.Type)
			merged = append(merged, lines...)
			merged = append(merged, comments...)

			annotation, err = parseLines(merged)
			if err != nil {
				return nil, err
			}
			annotation.Tags = names
		}
	}

	if err := annotation.Validate(); err != nil {
		return nil, err
	}
	return annotation, nil
}

// parseLines walks annotation comment lines in order, assigning as it goes.
// Later lines overwrite earlier ones. It does not run Validate.
func parseLines(comments []string) (*Annotation, error) {
	var annotation *Annotation

	for _, comment := range comments {
		// Strip "// " prefix and trim spaces
		clean := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(comment), "//"))
		if !strings.HasPrefix(clean, "@") {
			continue
		}

		parts := strings.Fields(clean)
		if len(parts) < 1 {
			continue
		}

		// Check for main marker: @temporal-gen-v2 activity|workflow
		if parts[0] == "@"+config.AnnotationPrefix {
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing type for %s annotation (expected 'activity' or 'workflow')", parts[0])
			}
			if annotation == nil {
				annotation = &Annotation{
					Type: parts[1],
				}
				if parts[1] == "activity" {
					annotation.ActivityOpts = &ActivityOptions{}
				} else if parts[1] == "workflow" {
					annotation.WorkflowOpts = &WorkflowOptions{}
				} else if parts[1] == "query" {
					annotation.QueryOpts = &QueryOptions{}
				} else if parts[1] == "signal" {
					annotation.SignalOpts = &SignalOptions{}
				} else if parts[1] == "update" {
					annotation.UpdateOpts = &UpdateOptions{}
				}
			}
			continue
		}

		// If we haven't found the main marker yet, ignore other flags
		if annotation == nil {
			continue
		}

		// Handle arguments
		switch parts[0] {
		// Common Arguments
		case "@tag":
			if len(parts) < 2 {
				return nil, tagErrorf("missing name for @tag (usage: @tag name)")
			}
			annotation.Tags = append(annotation.Tags, strings.Trim(parts[1], "\""))

		case "@id":
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @id")
			}
			if annotation.Type == "update" {
				annotation.UpdateOpts.ID = parts[1]
			}
			// TODO: Add other types that might support @id if needed

		// Activity Arguments
		case "@namespace":
			if annotation.Type != "activity" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @namespace")
			}
			annotation.ActivityOpts.Namespace = strings.Trim(parts[1], "\"")

		case "@task-queue":
			if annotation.Type != "activity" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @task-queue")
			}
			annotation.ActivityOpts.TaskQueue = strings.Trim(parts[1], "\"")

		case "@schedule-to-close-timeout":
			if annotation.Type != "activity" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @schedule-to-close-timeout")
			}
			d, err := time.ParseDuration(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid duration for @schedule-to-close-timeout: %w", err)
			}
			annotation.ActivityOpts.ScheduleToCloseTimeout = d

		case "@start-to-close-timeout":
			if annotation.Type != "activity" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @start-to-close-timeout")
			}
			d, err := time.ParseDuration(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid duration for @start-to-close-timeout: %w", err)
			}
			annotation.ActivityOpts.StartToCloseTimeout = d

		case "@schedule-to-start-timeout":
			if annotation.Type != "activity" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @schedule-to-start-timeout")
			}
			d, err := time.ParseDuration(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid duration for @schedule-to-start-timeout: %w", err)
			}
			annotation.ActivityOpts.ScheduleToStartTimeout = d

		case "@heartbeat-timeout":
			if annotation.Type != "activity" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @heartbeat-timeout")
			}
			d, err := time.ParseDuration(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid duration for @heartbeat-timeout: %w", err)
			}
			annotation.ActivityOpts.HeartbeatTimeout = d

		case "@wait-for-cancellation":
			if annotation.Type == "activity" {
				if len(parts) < 2 {
					return nil, fmt.Errorf("missing value for @wait-for-cancellation")
				}
				b, err := strconv.ParseBool(parts[1])
				if err != nil {
					return nil, fmt.Errorf("invalid boolean for @wait-for-cancellation: %w", err)
				}
				annotation.ActivityOpts.WaitForCancellation = b
			} else if annotation.Type == "workflow" {
				if len(parts) < 2 {
					return nil, fmt.Errorf("missing value for @wait-for-cancellation")
				}
				b, err := strconv.ParseBool(parts[1])
				if err != nil {
					return nil, fmt.Errorf("invalid boolean for @wait-for-cancellation: %w", err)
				}
				annotation.WorkflowOpts.WaitForCancellation = b
			}

		case "@activity-id":
			if annotation.Type != "activity" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @activity-id")
			}
			annotation.ActivityOpts.ActivityID = strings.Trim(parts[1], "\"")

		case "@disable-eager-execution":
			if annotation.Type != "activity" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @disable-eager-execution")
			}
			b, err := strconv.ParseBool(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid boolean for @disable-eager-execution: %w", err)
			}
			annotation.ActivityOpts.DisableEagerExecution = b

		case "@retry-policy-max-attempts":
			if annotation.Type != "activity" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @retry-policy-max-attempts")
			}
			n, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid number for @retry-policy-max-attempts: %w", err)
			}
			annotation.ActivityOpts.MaxRetries = n
			annotation.ActivityOpts.RetryPolicy = true

		case "@max-retries":
			// Kept for backward compatibility, same as @retry-policy-max-attempts
			if annotation.Type != "activity" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @max-retries")
			}
			n, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid number for @max-retries: %w", err)
			}
			annotation.ActivityOpts.MaxRetries = n
			annotation.ActivityOpts.RetryPolicy = true

		case "@options-callback":
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @options-callback")
			}
			val := strings.Trim(parts[1], "\"")
			if annotation.Type == "activity" {
				annotation.ActivityOpts.OptionsCallback = val
			} else if annotation.Type == "workflow" {
				annotation.WorkflowOpts.OptionsCallback = val
			}

		case "@by-field":
			if annotation.Type != "activity" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @by-field")
			}
			annotation.ActivityOpts.ByField = strings.Trim(parts[1], "\"")

		case "@by-field-only":
			if annotation.Type != "activity" {
				continue
			}
			annotation.ActivityOpts.ByFieldOnly = true

		case "@as-wrapper":
			if annotation.Type != "activity" {
				continue
			}
			annotation.ActivityOpts.GenerateWrapper = true

		case "@replica-read":
			if annotation.Type != "activity" {
				continue
			}
			annotation.ActivityOpts.ReplicaRead = true

		case "@local":
			if annotation.Type != "activity" {
				continue
			}
			annotation.ActivityOpts.IsLocal = true

		case "@wrapper-prefix":
			if annotation.Type != "activity" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @wrapper-prefix")
			}
			annotation.ActivityOpts.WrapperPrefix = strings.Trim(parts[1], "\"")

		// Workflow Arguments
		case "@execution-timeout":
			if annotation.Type != "workflow" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @execution-timeout")
			}
			d, err := time.ParseDuration(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid duration for @execution-timeout: %w", err)
			}
			annotation.WorkflowOpts.ExecutionTimeout = d

		case "@task-timeout":
			if annotation.Type != "workflow" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @task-timeout")
			}
			d, err := time.ParseDuration(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid duration for @task-timeout: %w", err)
			}
			annotation.WorkflowOpts.TaskTimeout = d

		case "@id-generator":
			if annotation.Type != "workflow" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @id-generator")
			}
			annotation.WorkflowOpts.IDGenerator = strings.Trim(parts[1], "\"")

		case "@id-template":
			if annotation.Type != "workflow" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @id-template")
			}
			tmplStr := strings.Join(parts[1:], " ")
			if _, err := template.New("workflowID").Parse(tmplStr); err != nil {
				return nil, fmt.Errorf("invalid template for @id-template: %w", err)
			}
			annotation.WorkflowOpts.IDTemplate = tmplStr

		case "@workflow-wait-for-cancellation":
			if annotation.Type != "workflow" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @workflow-wait-for-cancellation")
			}
			b, err := strconv.ParseBool(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid boolean for @workflow-wait-for-cancellation: %w", err)
			}
			annotation.WorkflowOpts.WaitForCancellation = b

		case "@workflow-task-queue":
			if annotation.Type != "workflow" {
				continue
			}
			if len(parts) < 2 {
				return nil, fmt.Errorf("missing value for @workflow-task-queue")
			}
			annotation.WorkflowOpts.TaskQueue = strings.Trim(parts[1], "\"")

		case "@memo":
			if annotation.Type != "workflow" {
				continue
			}
			if len(parts) < 3 {
				return nil, fmt.Errorf("missing key and value for @memo (usage: @memo key value)")
			}
			key := parts[1]
			value := strings.Trim(strings.Join(parts[2:], " "), "\"")
			if annotation.WorkflowOpts.Memo == nil {
				annotation.WorkflowOpts.Memo = make(map[string]string)
			}
			annotation.WorkflowOpts.Memo[key] = value

		default:
			// If it starts with @, assumes it's a directive. If we don't recognize it, error out.
			// We only error if it's inside a block we are parsing (annotation != nil)
			return nil, fmt.Errorf("unknown annotation argument: %s", parts[0])
		}
	}

	return annotation, nil
}
