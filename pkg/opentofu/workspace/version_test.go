package workspace

import (
	"os/exec"
	"testing"

	goversion "github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTofuVersion(t *testing.T) {
	tests := map[string]struct {
		out     string
		want    string
		wantErr string
	}{
		"plain banner": {
			out:  "OpenTofu v1.13.9\non linux_amd64\n",
			want: "1.13.9",
		},
		"banner only": {
			out:  "OpenTofu v1.6.2\n",
			want: "1.6.2",
		},
		"prerelease": {
			out:  "OpenTofu v1.9.0-alpha1\n",
			want: "1.9.0-alpha1",
		},
		"no banner": {
			out:     "tofu 1.13.9\n",
			wantErr: "unable to parse tofu version",
		},
		"empty": {
			out:     "",
			wantErr: "unable to parse tofu version",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			v, err := parseTofuVersion(tc.out)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			want, err := goversion.NewVersion(tc.want)
			require.NoError(t, err)
			assert.True(t, v.Equal(want), "got %s want %s", v, want)
		})
	}
}

func TestClientVersionCaches(t *testing.T) {
	client := newTofuClient("", "/bin/false", nil, nil)

	// The cache is version-probe-shaped: the first call runs `tofu version`
	// (and fails here, since /bin/false prints no banner), so exercise the
	// cache only after seeding it the way a successful run would.
	seeded, err := parseTofuVersion("OpenTofu v1.10.5\n")
	require.NoError(t, err)
	client.version = seeded

	got, err := client.Version(t.Context(), false)
	require.NoError(t, err)
	assert.True(t, got.Equal(seeded))
}

// TestVersionAgainstInstalledTofu runs the version probe against a real tofu
// binary when one is on the machine. It is skipped elsewhere; CI does not
// install tofu.
func TestVersionAgainstInstalledTofu(t *testing.T) {
	tofuPath, err := exec.LookPath("tofu")
	if err != nil {
		t.Skip("tofu not installed on this machine, skipping")
	}

	client := newTofuClient(t.TempDir(), tofuPath, nil, nil)

	v, err := client.Version(t.Context(), false)
	require.NoError(t, err)
	assert.NotNil(t, v)

	// The cached second call must return the same parsed version without
	// re-running the CLI.
	cached, err := client.Version(t.Context(), false)
	require.NoError(t, err)
	assert.True(t, v.Equal(cached))
}
