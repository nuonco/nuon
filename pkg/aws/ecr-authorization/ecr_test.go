package ecr

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	registryID := uuid.NewString()
	repository := fmt.Sprintf("%s.dkr.ecr.test", registryID)
	credentials := generics.GetFakeObj[*credentials.Config]()

	tests := map[string]struct {
		optFns      func() []Option
		assertFn    func(*testing.T, *ecrAuthorizer)
		errExpected error
	}{
		"happy path": {
			optFns: func() []Option {
				return []Option{
					WithRegistryID(registryID),
					WithCredentials(credentials),
				}
			},
			assertFn: func(t *testing.T, ecr *ecrAuthorizer) {
				assert.Equal(t, registryID, ecr.RegistryID)
				assert.Equal(t, credentials, ecr.Credentials)
			},
		},
		"happy path image": {
			optFns: func() []Option {
				return []Option{
					WithRegistryID(registryID),
					WithImageURL(fakeEcrImageURL),
					WithCredentials(credentials),
				}
			},
			assertFn: func(t *testing.T, ecr *ecrAuthorizer) {
				assert.NotEmpty(t, ecr.RegistryID)
				assert.Equal(t, credentials, ecr.Credentials)
			},
		},
		"happy path repository": {
			optFns: func() []Option {
				return []Option{
					WithCredentials(credentials),
					WithRepository(repository),
				}
			},
			assertFn: func(t *testing.T, ecr *ecrAuthorizer) {
				assert.NotEmpty(t, ecr.RegistryID)
				assert.Equal(t, credentials, ecr.Credentials)
			},
		},
		"happy path use default": {
			optFns: func() []Option {
				return []Option{
					WithUseDefault(true),
					WithCredentials(credentials),
				}
			},
			assertFn: func(t *testing.T, ecr *ecrAuthorizer) {
				assert.Equal(t, credentials, ecr.Credentials)
				assert.True(t, ecr.UseDefault)
			},
		},
		"invalid image": {
			optFns: func() []Option {
				return []Option{
					WithRegistryID(registryID),
					WithImageURL(uuid.NewString()),
					WithCredentials(credentials),
				}
			},
			errExpected: fmt.Errorf("invalid ecr image url"),
		},
		"not use default or image": {
			optFns: func() []Option {
				return []Option{
					WithCredentials(credentials),
				}
			},
			errExpected: fmt.Errorf("RegistryID"),
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
