package ecr

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	registryID := uuid.NewString()
	assumeRoleArn := uuid.NewString()
	assumeRoleSessionName := uuid.NewString()

	tests := map[string]struct {
		optFns      func() []Option
		assertFn    func(*testing.T, *ecrAuthorizer)
		errExpected error
	}{
		"happy path": {
			optFns: func() []Option {
				return []Option{
					WithRegistryID(registryID),
					WithAssumeRoleArn(assumeRoleArn),
					WithAssumeRoleSessionName(assumeRoleSessionName),
				}
			},
			assertFn: func(t *testing.T, ecr *ecrAuthorizer) {
				assert.Equal(t, registryID, ecr.RegistryID)
				assert.Equal(t, assumeRoleArn, ecr.AssumeRoleArn)
				assert.Equal(t, assumeRoleSessionName, ecr.AssumeRoleSessionName)
			},
		},
		"happy path image": {
			optFns: func() []Option {
				return []Option{
					WithRegistryID(registryID),
					WithImageURL(fakeEcrImageURL),
					WithAssumeRoleArn(assumeRoleArn),
					WithAssumeRoleSessionName(assumeRoleSessionName),
				}
			},
			assertFn: func(t *testing.T, ecr *ecrAuthorizer) {
				assert.NotEmpty(t, ecr.RegistryID)
				assert.Equal(t, assumeRoleArn, ecr.AssumeRoleArn)
				assert.Equal(t, assumeRoleSessionName, ecr.AssumeRoleSessionName)
			},
		},
		"invalid image": {
			optFns: func() []Option {
				return []Option{
					WithRegistryID(registryID),
					WithImageURL(uuid.NewString()),
					WithAssumeRoleArn(assumeRoleArn),
				}
			},
			errExpected: fmt.Errorf("invalid ecr image url"),
		},
		"missing assume role session name": {
			optFns: func() []Option {
				return []Option{
					WithRegistryID(registryID),
					WithAssumeRoleArn(assumeRoleArn),
				}
			},
			errExpected: fmt.Errorf("ecrAuthorizer.AssumeRoleSessionName"),
		},
		"missing assume role arn": {
			optFns: func() []Option {
				return []Option{
					WithRegistryID(registryID),
					WithAssumeRoleSessionName(assumeRoleSessionName),
				}
			},
			errExpected: fmt.Errorf("ecrAuthorizer.AssumeRoleArn"),
		},
		"missing registry id": {
			optFns: func() []Option {
				return []Option{
					WithAssumeRoleArn(assumeRoleArn),
					WithAssumeRoleSessionName(assumeRoleSessionName),
				}
			},
			errExpected: fmt.Errorf("ecrAuthorizer.RegistryID"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := validator.New()
			opts := test.optFns()
			ecr, err := New(v, opts...)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			test.assertFn(t, ecr)
		})
	}
}
