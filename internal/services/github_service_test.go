package services

import (
	"context"
	"errors"
	"testing"

	"github.com/powertoolsdev/api/internal/models"
	"github.com/stretchr/testify/assert"
)

func makePointerString(s string) *string {
	return &s
}

type mockRepoGetter struct {
	fakeGet func(int64) ([]*models.Repo, error)
}

func (m *mockRepoGetter) Repos(ctx context.Context, githubInstallationID int64) ([]*models.Repo, error) {
	if m.fakeGet != nil {
		return m.fakeGet(githubInstallationID)
	}

	return []*models.Repo{}, nil
}
func TestServiceRepos(t *testing.T) {
	tests := map[string]struct {
		input            int64
		mockRepoGetter   *mockRepoGetter
		expectedResponse []*models.Repo
		errExpected      error
	}{
		"happy path": {
			input: 1234567,
			mockRepoGetter: &mockRepoGetter{
				fakeGet: func(i int64) ([]*models.Repo, error) {
					return []*models.Repo{
						{URL: makePointerString("https://api.github.com/repos/octocat/Hello-World")},
					}, nil
				},
			},

			expectedResponse: []*models.Repo{
				{URL: makePointerString("https://api.github.com/repos/octocat/Hello-World")},
			},
		},
		"empty results": {
			input: 1234567,
			mockRepoGetter: &mockRepoGetter{
				fakeGet: func(i int64) ([]*models.Repo, error) {
					return []*models.Repo{}, nil
				},
			},
			expectedResponse: []*models.Repo{},
		},
		"error returns up through service": {
			input: 1234567,
			mockRepoGetter: &mockRepoGetter{
				fakeGet: func(i int64) ([]*models.Repo, error) {
					return []*models.Repo{}, errors.New("API error")
				},
			},
			errExpected: errors.New("API error"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			service := GithubService{
				repoGetter: test.mockRepoGetter,
			}

			success, _, err := service.Repos(context.Background(), test.input, &models.ConnectionOptions{})

			assert.Equal(t, test.expectedResponse, success)
			if test.errExpected != nil {
				assert.Error(t, err)
				assert.IsType(t, test.errExpected, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
