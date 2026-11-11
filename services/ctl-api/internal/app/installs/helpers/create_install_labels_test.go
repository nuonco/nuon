package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestCreateInstallLabels(t *testing.T) {
	cases := map[string]struct {
		parentApp      *app.App
		reqLabels      map[string]string
		wantMerged     map[string]string
		wantStatic     labels.Labels
		wantTemplated  labels.Labels
		wantErr        bool
		wantErrMessage string
	}{
		"app default labels fill in request labels when request has none": {
			parentApp:     &app.App{DefaultLabels: labels.Labels{"env": "prod"}},
			reqLabels:     nil,
			wantMerged:    map[string]string{"env": "prod"},
			wantStatic:    labels.Labels{"env": "prod"},
			wantTemplated: labels.Labels{},
		},
		"request label conflicting with app default label errors": {
			parentApp: &app.App{DefaultLabels: labels.Labels{"env": "prod"}},
			reqLabels: map[string]string{"env": "stage"},
			wantErr:   true,
		},
		"templated label is split out and not matched statically": {
			parentApp:     &app.App{},
			reqLabels:     map[string]string{"env": "{{ .nuon.install.name }}"},
			wantMerged:    map[string]string{"env": "{{ .nuon.install.name }}"},
			wantStatic:    labels.Labels{},
			wantTemplated: labels.Labels{"env": "{{ .nuon.install.name }}"},
		},
		"static and templated labels are both returned": {
			parentApp: &app.App{},
			reqLabels: map[string]string{
				"env":  "prod",
				"name": "{{ .nuon.install.name }}",
			},
			wantMerged: map[string]string{
				"env":  "prod",
				"name": "{{ .nuon.install.name }}",
			},
			wantStatic:    labels.Labels{"env": "prod"},
			wantTemplated: labels.Labels{"name": "{{ .nuon.install.name }}"},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			merged, static, templated, err := CreateInstallLabels(tc.parentApp, tc.reqLabels)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantMerged, merged)
			assert.Equal(t, tc.wantStatic, static)
			assert.Equal(t, tc.wantTemplated, templated)
		})
	}
}

func TestNoBranchForInstallErrorMentionsLabelsAndAppBranch(t *testing.T) {
	err := NoBranchForInstallError("acme-install")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "acme-install")
}
