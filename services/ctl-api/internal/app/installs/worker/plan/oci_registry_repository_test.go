package plan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImageNameSegment(t *testing.T) {
	cases := map[string]string{
		"img_altinity_clickhouse_operator": "img_altinity_clickhouse_operator",
		"My Component":                     "my-component",
		"foo   bar":                        "foo-bar",
		"foo--bar":                         "foo-bar",
		"-leading-and-trailing-":           "leading-and-trailing",
		"trailing---":                      "trailing",
		"...dots...":                       "dots",
		"CTL API (v2)":                     "ctl-api-v2",
		"":                                 "app",
		"---":                              "app",
	}

	for in, want := range cases {
		assert.Equal(t, want, imageNameSegment(in), "input %q", in)
	}
}
