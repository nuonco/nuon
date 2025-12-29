package infratests

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
)

const (
	TestSandboxWorkflowName = "TestSandbox"
)

type TestSandboxRequest struct {
	SandboxName string `json:"sandbox_name"`
}

func (req *TestSandboxRequest) Validate() error {
	return nil
}

type TestSandboxResponse struct {
	SandboxName string `json:"sandbox_name"`
}

func (req *TestSandboxResponse) Validate() error {
	return nil
}

func TestSandboxIDCallback(req *TestSandboxRequest) string {
	return fmt.Sprintf("%s-infra-test", req.SandboxName)
}

// @temporal-gen workflow
// @execution-timeout 1h
// @task-timeout 30m
// @task-queue "default"
// @namespace "infra-tests"
// @id-callback TestSandboxIDCallback
func TestSandbox(workflow.Context, *TestSandboxRequest) (*TestSandboxResponse, error) {
	panic("stub for code generation")
}
