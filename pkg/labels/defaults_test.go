package labels

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyDefaults(t *testing.T) {
	t.Run("applies static and templated defaults", func(t *testing.T) {
		newLabels, newTemplates, changed := ApplyDefaults(
			Labels{"env": "prod"},
			nil,
			nil,
			Labels{"tier": "gold", "region": "{{ .nuon.cloud_account.aws.region }}"},
		)
		require.True(t, changed)
		require.Equal(t, Labels{"env": "prod", "tier": "gold"}, newLabels)
		require.Equal(t, Labels{"region": "{{ .nuon.cloud_account.aws.region }}"}, newTemplates)
	})

	t.Run("removes defaults dropped from the app config", func(t *testing.T) {
		newLabels, newTemplates, changed := ApplyDefaults(
			Labels{"env": "prod", "tier": "gold", "region": "us-west-2"},
			Labels{"region": "{{ .nuon.cloud_account.aws.region }}"},
			Labels{"tier": "gold", "region": "{{ .nuon.cloud_account.aws.region }}"},
			nil,
		)
		require.True(t, changed)
		require.Equal(t, Labels{"env": "prod"}, newLabels)
		require.Empty(t, newTemplates)
	})

	t.Run("keeps user labels not in the snapshot", func(t *testing.T) {
		newLabels, _, changed := ApplyDefaults(
			Labels{"custom": "value", "tier": "gold"},
			nil,
			Labels{"tier": "gold"},
			Labels{"tier": "silver"},
		)
		require.True(t, changed)
		require.Equal(t, Labels{"custom": "value", "tier": "silver"}, newLabels)
	})

	t.Run("default switching from templated to static clears the template", func(t *testing.T) {
		newLabels, newTemplates, changed := ApplyDefaults(
			Labels{"region": "us-west-2"},
			Labels{"region": "{{ .nuon.cloud_account.aws.region }}"},
			Labels{"region": "{{ .nuon.cloud_account.aws.region }}"},
			Labels{"region": "static-region"},
		)
		require.True(t, changed)
		require.Equal(t, Labels{"region": "static-region"}, newLabels)
		require.Empty(t, newTemplates)
	})

	t.Run("no changes reports unchanged", func(t *testing.T) {
		_, _, changed := ApplyDefaults(
			Labels{"tier": "gold"},
			Labels{"region": "{{ .nuon.cloud_account.aws.region }}"},
			Labels{"tier": "gold", "region": "{{ .nuon.cloud_account.aws.region }}"},
			Labels{"tier": "gold", "region": "{{ .nuon.cloud_account.aws.region }}"},
		)
		require.False(t, changed)
	})
}
