package plan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImageNameSegment(t *testing.T) {
	cases := map[string]string{
		"img_nuon_ctl_api":                 "img-nuon-ctl-api",
		"img_altinity_clickhouse_operator": "img-altinity-clickhouse-operator",
		"My Component":                     "my-component",
		"foo   bar":                        "foo-bar",
		"foo--bar":                         "foo-bar",
		"-leading-and-trailing-":           "leading-and-trailing",
		"trailing___":                      "trailing",
		"...dots...":                       "dots",
		"CTL API (v2)":                     "ctl-api-v2",
		"":                                 "app",
		"___":                              "app",
	}

	for in, want := range cases {
		assert.Equal(t, want, imageNameSegment(in), "input %q", in)
	}
}
