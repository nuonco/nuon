package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClassifyResultContentCloudFormation(t *testing.T) {
	content := json.RawMessage(`{"stack_name":"demo","change_set_name":"candidate","changes":[]}`)
	require.Equal(t, "cloudformation", classifyResultContent(content))
}
