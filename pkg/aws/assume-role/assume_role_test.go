package iam

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()

	v := validator.New()
	tests := map[string]struct {
		v           *validator.Validate
		opts        []assumerOptions
		errExpected error
		expected    *assumer
	}{
		"happy path": {
			v: v,
			opts: []assumerOptions{
				WithRoleARN("valid:role-arn"),
				WithRoleSessionName("valid:session-name"),
			},
			expected: &assumer{RoleARN: "valid:role-arn", RoleSessionName: "valid:session-name"},
		},
		"happy path with settings": {
			v: v,
			opts: []assumerOptions{
				WithSettings(Settings{
					RoleARN:         "valid:role-arn",
					RoleSessionName: "valid:session-name",
				}),
			},
			expected: &assumer{RoleARN: "valid:role-arn", RoleSessionName: "valid:session-name"},
		},
		"invalid settings": {
			v: v,
			opts: []assumerOptions{
				WithSettings(Settings{
					RoleSessionName: "valid:session-name",
				}),
			},
			errExpected: fmt.Errorf("RoleARN"),
		},
		"missing validator": {
			opts: []assumerOptions{
				WithRoleARN("valid:role:arn"),
				WithRoleSessionName("valid-session-name"),
			},
			errExpected: fmt.Errorf("validator is nil"),
		},
		"no options": {
			v:           v,
			opts:        []assumerOptions{},
			errExpected: fmt.Errorf("Field validation"),
		},
		"missing arn": {
			v: v,
			opts: []assumerOptions{
				WithRoleSessionName("valid-session-name"),
			},
			errExpected: fmt.Errorf("Field validation for 'RoleARN'"),
		},
		"missing role session name": {
			v: v,
			opts: []assumerOptions{
				WithRoleARN("valid:role:arn"),
			},
			errExpected: fmt.Errorf("Field validation for 'RoleSessionName'"),
		},
	}

	for name, test := range tests {
		name := name
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			a, err := New(test.v, test.opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expected.RoleARN, a.RoleARN)
			assert.Equal(t, test.expected.RoleSessionName, a.RoleSessionName)
		})
	}
}
