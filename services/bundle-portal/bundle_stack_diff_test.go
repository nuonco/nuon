package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/runner/customer_managed/operation"
)

func TestAddStackPropertyChanges(t *testing.T) {
	candidate := &operation.StackCandidate{Changes: []operation.StackChange{
		{LogicalResourceID: "Runner"},
		{LogicalResourceID: "UnchangedDependency"},
	}}
	before := []byte(`{"Resources":{"Runner":{"Properties":{"Image":"runner:v1","Tags":[{"Key":"Name","Value":"runner"},{"Key":"Version","Value":"v1"}]}},"UnchangedDependency":{"Properties":{"Name":"same"}}}}`)
	after := []byte(`{"Resources":{"Runner":{"Properties":{"Image":"runner:v2","Tags":[{"Key":"Version","Value":"v2"},{"Key":"Name","Value":"runner"}]}},"UnchangedDependency":{"Properties":{"Name":"same"}}}}`)

	require.NoError(t, addStackPropertyChanges(candidate, before, after))
	require.Equal(t, []operation.StackPropertyChange{
		{Path: "Properties.Image", Before: "runner:v1", After: "runner:v2"},
		{Path: "Properties.Tags[Version].Value", Before: "v1", After: "v2"},
	}, candidate.Changes[0].PropertyChanges)
	require.Empty(t, candidate.Changes[1].PropertyChanges)
}

func TestStackCandidateTemplateKey(t *testing.T) {
	require.Equal(t, "stack/candidates/sha256-abc/stack/root-template.json", stackCandidateTemplateKey("sha256:abc"))
}

func TestStackPropertyChangesCaptured(t *testing.T) {
	require.False(t, stackPropertyChangesCaptured(operation.StackCandidate{Changes: []operation.StackChange{{LogicalResourceID: "Runner"}}}))
	require.True(t, stackPropertyChangesCaptured(operation.StackCandidate{Changes: []operation.StackChange{{PropertyChanges: []operation.StackPropertyChange{{Path: "Properties.Image"}}}}}))
	require.True(t, stackPropertyChangesCaptured(operation.StackCandidate{Changes: []operation.StackChange{{PropertyChangesTruncated: true}}}))
}
