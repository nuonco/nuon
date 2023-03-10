package github

import (
	"context"
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/google/go-github/v50/github"
	"github.com/google/uuid"
	"github.com/powertoolsdev/go-generics"
	"github.com/stretchr/testify/assert"
)

func Test_createInstallationToken(t *testing.T) {
	testErr := fmt.Errorf("testErr")
	repoName := uuid.NewString()
	token := uuid.NewString()
	installID := int64(987654321)

	tests := map[string]struct {
		clientFn    func(*gomock.Controller) installationTokenCreatorClient
		assertFn    func(*testing.T, string)
		errExpected error
	}{
		"happy path": {
			clientFn: func(mockCtl *gomock.Controller) installationTokenCreatorClient {
				mock := NewMockinstallationTokenCreatorClient(mockCtl)
				opts := &github.InstallationTokenOptions{
					Repositories: []string{repoName},
				}
				resp := &github.InstallationToken{
					Token: generics.ToPtr(token),
					Repositories: []*github.Repository{
						{
							Name: generics.ToPtr(repoName),
						},
					},
				}
				mock.EXPECT().CreateInstallationToken(gomock.Any(), installID, opts).Return(resp, nil, nil)
				return mock
			},
			assertFn: func(t *testing.T, respToken string) {
				assert.Equal(t, token, respToken)
			},
		},
		"client err": {
			clientFn: func(mockCtl *gomock.Controller) installationTokenCreatorClient {
				mock := NewMockinstallationTokenCreatorClient(mockCtl)
				opts := &github.InstallationTokenOptions{
					Repositories: []string{repoName},
				}
				mock.EXPECT().CreateInstallationToken(gomock.Any(), installID, opts).Return(nil, nil, testErr)
				return mock
			},
			errExpected: testErr,
		},
		"no permissions granted": {
			clientFn: func(mockCtl *gomock.Controller) installationTokenCreatorClient {
				mock := NewMockinstallationTokenCreatorClient(mockCtl)
				opts := &github.InstallationTokenOptions{
					Repositories: []string{repoName},
				}
				resp := &github.InstallationToken{
					Token: generics.ToPtr(token),
					Repositories: []*github.Repository{
						{
							Name: generics.ToPtr(uuid.NewString()),
						},
					},
				}
				mock.EXPECT().CreateInstallationToken(gomock.Any(), installID, opts).Return(resp, nil, nil)
				return mock
			},
			errExpected: fmt.Errorf("installation does not allow"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			mockCtl := gomock.NewController(t)
			ghClient := test.clientFn(mockCtl)

			g := &gh{
				RepoName:  repoName,
				InstallID: installID,
			}
			token, err := g.createInstallationToken(ctx, ghClient)
			if test.errExpected != nil {
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}

			assert.NoError(t, err)
			test.assertFn(t, token)
		})
	}
}
