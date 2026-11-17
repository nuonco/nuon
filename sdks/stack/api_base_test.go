package stack

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIBase(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{in: "https://runner.nuon.co", want: "https://stack.nuon.co"},
		{in: "https://runner.stage.nuon.co/", want: "https://stack.stage.nuon.co"},
		{in: "https://runner.stage.nuon.co/v1/stacks/old/config", want: "https://stack.stage.nuon.co"},
		{in: "https://stack.stage.nuon.co", want: "https://stack.stage.nuon.co"},
		{in: "http://127.0.0.1:8083/runner", want: "http://127.0.0.1:8083"},
	}
	for _, tc := range cases {
		got, err := APIBase(tc.in)
		require.NoError(t, err)
		require.Equal(t, tc.want, got, tc.in)
	}
}
