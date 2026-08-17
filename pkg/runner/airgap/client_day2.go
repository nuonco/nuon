package airgap

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func newGzipReader(b []byte) (*gzip.Reader, error) {
	return gzip.NewReader(bytes.NewReader(b))
}
func readAll(r io.Reader) ([]byte, error) { return io.ReadAll(r) }

type RuntimeJob struct {
	RuntimeJobID  string
	JobType       string
	JobGroup      string
	JobOperation  string
	CompositePlan json.RawMessage
	Envelope      *Envelope
	runID         string
	templateID    string
	started       bool
	completed     bool
	success       bool
	err           string
	result        *models.ServiceCreateRunnerJobExecutionResultRequest
	outputs       *models.ServiceCreateRunnerJobExecutionOutputsRequest
	done          chan struct{}
}

type Day2JobHandle struct{ job *RuntimeJob }

func (h *Day2JobHandle) Await(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-h.job.done:
		if !h.job.success {
			return fmt.Errorf("day-2 job failed: %s", h.job.err)
		}
		return nil
	}
}

func (h *Day2JobHandle) Result() *models.ServiceCreateRunnerJobExecutionResultRequest {
	return h.job.result
}
func (h *Day2JobHandle) Outputs() *models.ServiceCreateRunnerJobExecutionOutputsRequest {
	return h.job.outputs
}

func (h *Day2JobHandle) PlanJSON() ([]byte, error) {
	if h.job.result == nil || h.job.result.ContentsDisplayCompressed == "" {
		return nil, fmt.Errorf("day-2 job has no terraform plan JSON")
	}
	compressed, err := base64.URLEncoding.DecodeString(h.job.result.ContentsDisplayCompressed)
	if err != nil {
		return nil, err
	}
	reader, err := newGzipReader(compressed)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return readAll(reader)
}

func (c *Client) EnqueueDay2Job(runtimeJobID, jobType, jobGroup, jobOperation string, compositePlan json.RawMessage) (*Day2JobHandle, error) {
	return c.EnqueueDay2JobWithEnvelope(runtimeJobID, jobType, jobGroup, jobOperation, compositePlan, c.envelope)
}

func (c *Client) EnqueueDay2JobWithEnvelope(runtimeJobID, jobType, jobGroup, jobOperation string, compositePlan json.RawMessage, envelope *Envelope) (*Day2JobHandle, error) {
	runID, templateID, ok := strings.Cut(runtimeJobID, "--")
	if !ok || runID == "" || templateID == "" {
		return nil, fmt.Errorf("runtime job ID %q must be <run-id>--<template-id>", runtimeJobID)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.result.Succeeded {
		return nil, fmt.Errorf("bootstrap is not complete")
	}
	if _, exists := c.runtimeJobs[runtimeJobID]; exists {
		return nil, fmt.Errorf("runtime job %q already exists", runtimeJobID)
	}
	job := &RuntimeJob{RuntimeJobID: runtimeJobID, JobType: jobType, JobGroup: jobGroup, JobOperation: jobOperation, CompositePlan: compositePlan, Envelope: envelope, runID: runID, templateID: templateID, done: make(chan struct{})}
	c.runtimeJobs[runtimeJobID] = job
	c.runtimeQueue = append(c.runtimeQueue, runtimeJobID)
	return &Day2JobHandle{job: job}, nil
}

func (c *Client) getRuntimeJob(id string) *RuntimeJob {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.runtimeJobs[id]
}

func (c *Client) runtimeActionJob(runID string) *RuntimeJob {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, job := range c.runtimeJobs {
		if job.runID == runID || job.RuntimeJobID == runID {
			return job
		}
		var composite struct {
			ActionWorkflowRunPlan *struct {
				ID string `json:"id"`
			} `json:"action_workflow_run_plan"`
		}
		if json.Unmarshal(job.CompositePlan, &composite) == nil && composite.ActionWorkflowRunPlan != nil && composite.ActionWorkflowRunPlan.ID == runID {
			return job
		}
	}
	return nil
}

func (c *Client) runtimeJob(j *RuntimeJob) *models.AppRunnerJob {
	return &models.AppRunnerJob{ID: j.RuntimeJobID, Type: models.AppRunnerJobType(j.JobType), Operation: models.AppRunnerJobOperationType(j.JobOperation), Group: models.AppRunnerJobGroup(j.JobGroup), Status: models.AppRunnerJobStatusAvailable, LogStreamID: "airgap-" + j.RuntimeJobID, OrgID: c.envelope.OrgID, OwnerID: c.envelope.InstallID, ExecutionTimeout: int64(24 * time.Hour)}
}

func (c *Client) writeRuntimeJSON(job *RuntimeJob, name string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return c.store.WriteFile(day2.RunsPrefix+job.runID+"/steps/"+job.templateID+"/"+name, append(b, '\n'))
}

func (c *Client) completeRuntimeLocked(job *RuntimeJob, success bool, message string) {
	if job.completed {
		return
	}
	job.completed, job.success, job.err = true, success, message
	for i, id := range c.runtimeQueue {
		if id == job.RuntimeJobID {
			c.runtimeQueue = append(c.runtimeQueue[:i], c.runtimeQueue[i+1:]...)
			break
		}
	}
	close(job.done)
}
