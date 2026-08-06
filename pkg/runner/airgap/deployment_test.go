package airgap

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeploymentInstallID(t *testing.T) {
	const frozen = "inlqwvpl9h932fzsjpyqedzn93"

	derived, err := DeploymentInstallID(frozen, "demo2")
	require.NoError(t, err)
	require.Equal(t, "inlqwvpl9h932fzsjpyq-demo2", derived)
	require.Len(t, derived, len(frozen))

	again, err := DeploymentInstallID(frozen, "demo2")
	require.NoError(t, err)
	require.Equal(t, derived, again)

	other, err := DeploymentInstallID(frozen, "demo3")
	require.NoError(t, err)
	require.NotEqual(t, derived, other)
	require.Len(t, other, len(frozen))

	longest, err := DeploymentInstallID(frozen, "abcd1234")
	require.NoError(t, err)
	require.Equal(t, "inlqwvpl9h932fzsj-abcd1234", longest)
}

func TestDeploymentInstallIDRejectsInvalidSuffixes(t *testing.T) {
	const frozen = "inlqwvpl9h932fzsjpyqedzn93"
	for _, suffix := range []string{"", "Demo2", "has-dash", "under_x", "toolong99", "d mo"} {
		_, err := DeploymentInstallID(frozen, suffix)
		require.Error(t, err, "suffix %q", suffix)
	}
}

func TestDeploymentInstallIDRejectsShortInstallID(t *testing.T) {
	_, err := DeploymentInstallID("inlshort", "demo2")
	require.Error(t, err)

	_, err = DeploymentInstallID("inl1234567890abcde", "demo2")
	require.NoError(t, err)

	_, err = DeploymentInstallID("inl1234567890abcd", "demo2")
	require.Error(t, err)
}

func TestApplyDeploymentID(t *testing.T) {
	const frozen = "inlqwvpl9h932fzsjpyqedzn93"
	plan := `{"role_arn":"arn:aws:iam::111122223333:role/` + frozen + `-provision","cluster":"w-` + frozen + `","install_id":"` + frozen + `","other":"unrelated-inlsomethingelse"}`
	envelope := &Envelope{
		Version:   "v0",
		InstallID: frozen,
		AppConfig: json.RawMessage(`{"name":"` + frozen + `"}`),
		Steps: []Step{
			{ID: "step-1", CompositePlan: json.RawMessage(plan)},
			{ID: "step-2", CompositePlan: json.RawMessage(`{"no_id_here":true}`)},
		},
	}

	derived, err := envelope.ApplyDeploymentID("demo2")
	require.NoError(t, err)
	require.Equal(t, "inlqwvpl9h932fzsjpyq-demo2", derived)
	require.Equal(t, derived, envelope.InstallID)

	rewritten := string(envelope.Steps[0].CompositePlan)
	require.NotContains(t, rewritten, frozen)
	require.Contains(t, rewritten, "arn:aws:iam::111122223333:role/"+derived+"-provision")
	require.Contains(t, rewritten, `"cluster":"w-`+derived+`"`)
	require.Contains(t, rewritten, "unrelated-inlsomethingelse")
	require.JSONEq(t, `{"no_id_here":true}`, string(envelope.Steps[1].CompositePlan))
	require.JSONEq(t, `{"name":"`+derived+`"}`, string(envelope.AppConfig))
}

func TestApplyDeploymentIDInvalidSuffixLeavesEnvelopeUntouched(t *testing.T) {
	const frozen = "inlqwvpl9h932fzsjpyqedzn93"
	envelope := &Envelope{
		Version:   "v0",
		InstallID: frozen,
		Steps:     []Step{{ID: "step-1", CompositePlan: json.RawMessage(`{"install_id":"` + frozen + `"}`)}},
	}
	_, err := envelope.ApplyDeploymentID("NOPE")
	require.Error(t, err)
	require.Equal(t, frozen, envelope.InstallID)
	require.Contains(t, string(envelope.Steps[0].CompositePlan), frozen)
}
