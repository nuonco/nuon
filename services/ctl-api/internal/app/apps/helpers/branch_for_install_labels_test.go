package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestBranchSelectsLabels(t *testing.T) {
	cases := map[string]struct {
		groups []app.AppBranchInstallGroup
		lbls   labels.Labels
		match  bool
	}{
		"no groups does not match": {
			groups: nil,
			lbls:   labels.Labels{"env": "prod"},
			match:  false,
		},
		"default group with no selector does not match labeled install": {
			groups: []app.AppBranchInstallGroup{{Name: "default", Default: true}},
			lbls:   labels.Labels{"env": "prod"},
			match:  false,
		},
		"default group with no selector does not match unlabeled install": {
			groups: []app.AppBranchInstallGroup{{Name: "default", Default: true}},
			lbls:   nil,
			match:  false,
		},
		"default group selecting env=prod matches": {
			groups: []app.AppBranchInstallGroup{{
				Name:          "default",
				Default:       true,
				LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}},
			}},
			lbls:  labels.Labels{"env": "prod"},
			match: true,
		},
		"non-default group selecting env=prod plus unlabeled default matches": {
			groups: []app.AppBranchInstallGroup{
				{Name: "prod", LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}}},
				{Name: "default", Default: true},
			},
			lbls:  labels.Labels{"env": "prod"},
			match: true,
		},
		"group selecting env=prod does not match env=stage": {
			groups: []app.AppBranchInstallGroup{{
				Name:          "prod",
				LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}},
			}},
			lbls:  labels.Labels{"env": "stage"},
			match: false,
		},
		"group selecting env=prod,tier=gold does not match only env=prod": {
			groups: []app.AppBranchInstallGroup{{
				Name:          "prod-gold",
				LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod", "tier": "gold"}},
			}},
			lbls:  labels.Labels{"env": "prod"},
			match: false,
		},
		"group selecting env=prod,tier=gold matches superset of labels": {
			groups: []app.AppBranchInstallGroup{{
				Name:          "prod-gold",
				LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod", "tier": "gold"}},
			}},
			lbls:  labels.Labels{"env": "prod", "tier": "gold", "team": "x"},
			match: true,
		},
		"group selecting env=* matches any env value": {
			groups: []app.AppBranchInstallGroup{{
				Name:          "any-env",
				LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "*"}},
			}},
			lbls:  labels.Labels{"env": "anything"},
			match: true,
		},
		"group selecting env=prod excludes region=eu via not_match_labels": {
			groups: []app.AppBranchInstallGroup{{
				Name: "prod-not-eu",
				LabelSelector: &labels.Selector{
					MatchLabels:    labels.Labels{"env": "prod"},
					NotMatchLabels: labels.Labels{"region": "eu"},
				},
			}},
			lbls:  labels.Labels{"env": "prod", "region": "eu"},
			match: false,
		},
		"group with only not_match_labels never matches": {
			groups: []app.AppBranchInstallGroup{{
				Name:          "not-eu",
				LabelSelector: &labels.Selector{NotMatchLabels: labels.Labels{"region": "eu"}},
			}},
			lbls:  labels.Labels{"region": "us"},
			match: false,
		},
		"any groups with no labels never matches": {
			groups: []app.AppBranchInstallGroup{
				{Name: "prod", LabelSelector: &labels.Selector{MatchLabels: labels.Labels{"env": "prod"}}},
				{Name: "default", Default: true},
			},
			lbls:  nil,
			match: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.match, branchSelectsLabels(tc.groups, tc.lbls))
		})
	}
}
