package sessionpolicy

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepositoryARN(t *testing.T) {
	arn, err := RepositoryARN("111122223333.dkr.ecr.us-west-2.amazonaws.com/org1/app1")
	require.NoError(t, err)
	require.Equal(t, "arn:aws:ecr:us-west-2:111122223333:repository/org1/app1", arn)

	_, err = RepositoryARN("registry.example.com/org/app")
	require.Error(t, err)
	_, err = RepositoryARN("111122223333.dkr.ecr.us-west-2.amazonaws.com/org/*")
	require.Error(t, err)
}

func TestDocumentLimitsTheTokenToOneRepository(t *testing.T) {
	doc, err := Document("arn:aws:ecr:us-west-2:111122223333:repository/org1/app1", true)
	require.NoError(t, err)

	var parsed struct {
		Statement []struct {
			Action   any    `json:"Action"`
			Resource string `json:"Resource"`
		} `json:"Statement"`
	}
	require.NoError(t, json.Unmarshal([]byte(doc), &parsed))
	require.Equal(t, "*", parsed.Statement[0].Resource)
	require.Equal(t, "arn:aws:ecr:us-west-2:111122223333:repository/org1/app1", parsed.Statement[1].Resource)
	require.Contains(t, doc, "ecr:PutImage")

	pullOnly, err := Document("arn:aws:ecr:us-west-2:111122223333:repository/org1/app1", false)
	require.NoError(t, err)
	require.NotContains(t, pullOnly, "ecr:PutImage")
	require.True(t, strings.Contains(pullOnly, "ecr:BatchGetImage"))
}
