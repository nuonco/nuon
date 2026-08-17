package bundleupgrade

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
)

type memoryStore struct {
	objects map[string][]byte
	writes  []string
}

func (m *memoryStore) GetObject(_ context.Context, in *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	raw, ok := m.objects[*in.Key]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(raw))}, nil
}

func (m *memoryStore) PutObject(_ context.Context, in *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	raw, err := io.ReadAll(in.Body)
	if err != nil {
		return nil, err
	}
	m.writes = append(m.writes, *in.Key)
	m.objects[*in.Key] = raw
	return &s3.PutObjectOutput{}, nil
}

func TestValidateIdentity(t *testing.T) {
	active := day2.BundleInfo{DeploymentID: "install-1", BundleDigest: "sha256:active"}
	require.NoError(t, validateIdentity(active, "install-1", "sha256:active", "install-1", "install-1"))
	require.ErrorContains(t, validateIdentity(active, "install-1", "sha256:active", "install-other", "install-1"), "status install ID")
	require.ErrorContains(t, validateIdentity(active, "install-other", "sha256:active", "install-1", "install-1"), "catalog deployment ID")
	require.ErrorContains(t, validateIdentity(active, "install-1", "sha256:other", "install-1", "install-1"), "catalog bundle digest")
}

func TestNeedsCanonicalHydration(t *testing.T) {
	require.True(t, needsCanonicalHydration(day2.BundleInfo{Contents: []day2.BundleContent{{Kind: day2.BundleContentKindAction, Name: "restart"}}}))
	require.True(t, needsCanonicalHydration(day2.BundleInfo{Contents: []day2.BundleContent{{Kind: day2.BundleContentKindComponent, Name: "api"}}}))
	require.False(t, needsCanonicalHydration(day2.BundleInfo{Contents: []day2.BundleContent{{Kind: day2.BundleContentKindAction, Name: "restart", ActionDefinition: &day2.BundleActionDefinition{}}}}))
}

func TestStageCandidateWritesDigestScopedArchiveThenControlRecords(t *testing.T) {
	archive := t.TempDir() + "/candidate.tar.zst"
	require.NoError(t, os.WriteFile(archive, []byte("archive"), 0o600))
	store := &memoryStore{objects: map[string][]byte{}}
	record := day2.CandidateStageKey("sha256:next", fixedTime())
	var progress []Progress
	err := stageCandidate(context.Background(), store, "bucket", "state/control/", archive, "bundle/candidates/sha256-next.tar.zst", record, []byte(`{"bundle":{}}`), func(update Progress) {
		progress = append(progress, update)
	})
	require.NoError(t, err)
	require.Equal(t, []Progress{
		{Phase: "publishing-archive", Detail: "Publishing the candidate bundle archive"},
		{Phase: "recording", Detail: "Recording the staged bundle and its diff"},
	}, progress)
	require.Equal(t, []string{
		"bundle/candidates/sha256-next.tar.zst",
		"state/control/" + record,
		"state/control/" + StagedCandidateKey,
	}, store.writes)
	require.Equal(t, store.objects["state/control/"+record], store.objects["state/control/"+StagedCandidateKey])
}

func fixedTime() time.Time {
	return time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)
}
