package repos

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	gh "github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v41/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

var key = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEA0BUezcR7uycgZsfVLlAf4jXP7uFpVh4geSTY39RvYrAll0yh
q7uiQypP2hjQJ1eQXZvkAZx0v9lBYJmX7e0HiJckBr8+/O2kARL+GTCJDJZECpjy
97yylbzGBNl3s76fZ4CJ+4f11fCh7GJ3BJkMf9NFhe8g1TYS0BtSd/sauUQEuG/A
3fOJxKTNmICZr76xavOQ8agA4yW9V5hKcrbHzkfecg/sQsPMmrXixPNxMsqyOMmg
jdJ1aKr7ckEhd48ft4bPMO4DtVL/XFdK2wJZZ0gXJxWiT1Ny41LVql97Odm+OQyx
tcayMkGtMb1nwTcVVl+RG2U5E1lzOYpcQpyYFQIDAQABAoIBAAfUY55WgFlgdYWo
i0r81NZMNBDHBpGo/IvSaR6y/aX2/tMcnRC7NLXWR77rJBn234XGMeQloPb/E8iw
vtjDDH+FQGPImnQl9P/dWRZVjzKcDN9hNfNAdG/R9JmGHUz0JUddvNNsIEH2lgEx
C01u/Ntqdbk+cDvVlwuhm47MMgs6hJmZtS1KDPgYJu4IaB9oaZFN+pUyy8a1w0j9
RAhHpZrsulT5ThgCra4kKGDNnk2yfI91N9lkP5cnhgUmdZESDgrAJURLS8PgInM4
YPV9L68tJCO4g6k+hFiui4h/4cNXYkXnaZSBUoz28ICA6e7I3eJ6Y1ko4ou+Xf0V
csM8VFkCgYEA7y21JfECCfEsTHwwDg0fq2nld4o6FkIWAVQoIh6I6o6tYREmuZ/1
s81FPz/lvQpAvQUXGZlOPB9eW6bZZFytcuKYVNE/EVkuGQtpRXRT630CQiqvUYDZ
4FpqdBQUISt8KWpIofndrPSx6JzI80NSygShQsScWFw2wBIQAnV3TpsCgYEA3reL
L7AwlxCacsPvkazyYwyFfponblBX/OvrYUPPaEwGvSZmE5A/E4bdYTAixDdn4XvE
ChwpmRAWT/9C6jVJ/o1IK25dwnwg68gFDHlaOE+B5/9yNuDvVmg34PWngmpucFb/
6R/kIrF38lEfY0pRb05koW93uj1fj7Uiv+GWRw8CgYEAn1d3IIDQl+kJVydBKItL
tvoEur/m9N8wI9B6MEjhdEp7bXhssSvFF/VAFeQu3OMQwBy9B/vfaCSJy0t79uXb
U/dr/s2sU5VzJZI5nuDh67fLomMni4fpHxN9ajnaM0LyI/E/1FFPgqM+Rzb0lUQb
yqSM/ptXgXJls04VRl4VjtMCgYEAprO/bLx2QjxdPpXGFcXbz6OpsC92YC2nDlsP
3cfB0RFG4gGB2hbX/6eswHglLbVC/hWDkQWvZTATY2FvFps4fV4GrOt5Jn9+rL0U
elfC3e81Dw+2z7jhrE1ptepprUY4z8Fu33HNcuJfI3LxCYKxHZ0R2Xvzo+UYSBqO
ng0eTKUCgYEAxW9G4FjXQH0bjajntjoVQGLRVGWnteoOaQr/cy6oVii954yNMKSP
rezRkSNbJ8cqt9XQS+NNJ6Xwzl3EbuAt6r8f8VO1TIdRgFOgiUXRVNZ3ZyW8Hegd
kGTL0A6/0yAu9qQZlFbaD5bWhQo7eyx63u4hZGppBhkTSPikOYUPCH8=
-----END RSA PRIVATE KEY-----`)

func toPtr[T any](t T) *T {
	return &t
}

func TestRepos(t *testing.T) {
	tests := map[string]struct {
		input            string
		githubClient     *http.Client
		expectedResponse []*models.Repo
		errExpected      error
	}{
		"token fails": {
			input: "1234567",
			githubClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostAppInstallationsAccessTokensByInstallationId,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						mock.WriteError(
							w,
							http.StatusUnauthorized,
							"tokenError",
						)
					}),
				),
			),
			errExpected: errors.New("received non 2xx response status \"401 Unauthorized\""),
		},
		"happy path with results": {
			input: "1234567",
			githubClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.PostAppInstallationsAccessTokensByInstallationId,
					struct {
						Token     string    `json:"token"`
						ExpiresAt time.Time `json:"expires_at"`
					}{
						Token:     "this-is-the-token",
						ExpiresAt: time.Now().Add(5 * time.Minute),
					},
				),
				mock.WithRequestMatchHandler(
					mock.GetInstallationRepositories,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						_, err := w.Write(mock.MustMarshal(&github.ListRepositories{
							TotalCount: toPtr(1),
							Repositories: []*github.Repository{
								{
									ID:            toPtr(int64(1296269)),
									NodeID:        toPtr("MDEwOlJlcG9zaXRvcnkxMjk2MjY5"),
									Name:          toPtr("Hello-World"),
									FullName:      toPtr("octocat/Hello-World"),
									HTMLURL:       toPtr("https://api.github.com/repos/octocat/Hello-World"),
									DefaultBranch: toPtr("master"),
									Owner: &github.User{
										Login: toPtr("octocat"),
									},
									Private: toPtr(true),
								},
							},
						}))
						assert.NoError(t, err)
					}),
				),
			),
			expectedResponse: []*models.Repo{
				{
					URL:           toPtr("https://api.github.com/repos/octocat/Hello-World"),
					DefaultBranch: toPtr("master"),
					Owner:         toPtr("octocat"),
					Name:          toPtr("Hello-World"),
					FullName:      toPtr("octocat/Hello-World"),
					Private:       toPtr(true),
				},
			},
			errExpected: nil,
		},
		"empty results works": {
			input: "1234567",
			githubClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.PostAppInstallationsAccessTokensByInstallationId,
					struct {
						Token     string    `json:"token"`
						ExpiresAt time.Time `json:"expires_at"`
					}{
						Token:     "this-is-the-token",
						ExpiresAt: time.Now().Add(5 * time.Minute),
					},
				),
				mock.WithRequestMatchHandler(
					mock.GetInstallationRepositories,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						_, err := w.Write(mock.MustMarshal(&github.ListRepositories{
							TotalCount:   toPtr(0),
							Repositories: []*github.Repository{},
						}))
						assert.NoError(t, err)
					}),
				),
			),
			expectedResponse: []*models.Repo{},
			errExpected:      nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			appTransport, err := gh.NewAppsTransport(test.githubClient.Transport, 123456, key)
			require.NoError(t, err)

			githubRepo := NewGithubRepo(appTransport, zaptest.NewLogger(t), test.githubClient)

			r, err := githubRepo.Repos(context.Background(), test.input)

			if test.errExpected != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResponse, r)
		})
	}
}

func TestGetCommit(t *testing.T) {
	tests := map[string]struct {
		githubInstallID  string
		githubRepoOwner  string
		githubRepo       string
		githubBranch     string
		githubClient     *http.Client
		expectedResponse *github.RepositoryCommit
		errExpected      error
	}{
		"auth error": {
			githubInstallID: "1234567",
			githubRepoOwner: "octocat",
			githubRepo:      "Hello-World",
			githubBranch:    "master",
			githubClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostAppInstallationsAccessTokensByInstallationId,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						mock.WriteError(
							w,
							http.StatusUnauthorized,
							"tokenError",
						)
					}),
				),
			),
			errExpected: errors.New("received non 2xx response status \"401 Unauthorized\""),
		},
		"happy path": {
			githubInstallID: "1234567",
			githubRepoOwner: "octocat",
			githubRepo:      "Hello-World",
			githubBranch:    "master",
			githubClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.PostAppInstallationsAccessTokensByInstallationId,
					struct {
						Token     string    `json:"token"`
						ExpiresAt time.Time `json:"expires_at"`
					}{
						Token:     "this-is-the-token",
						ExpiresAt: time.Now().Add(5 * time.Minute),
					},
				),
				mock.WithRequestMatchHandler(
					mock.GetReposCommitsByOwnerByRepoByRef,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						_, err := w.Write(mock.MustMarshal(&github.RepositoryCommit{
							Commit: &github.Commit{
								SHA:     toPtr("9d13991a7d8cde95817fc2bf50380437dbec79ab"),
								Message: toPtr("fix stuff"),
								URL:     toPtr("https://github.com/octocat/Hello-World/commit/9d13991a7d8cde95817fc2bf50380437dbec79ab"),
							},
						}))
						assert.NoError(t, err)
					}),
				),
			),
			expectedResponse: &github.RepositoryCommit{
				Commit: &github.Commit{
					SHA:     toPtr("9d13991a7d8cde95817fc2bf50380437dbec79ab"),
					Message: toPtr("fix stuff"),
					URL:     toPtr("https://github.com/octocat/Hello-World/commit/9d13991a7d8cde95817fc2bf50380437dbec79ab"),
				},
			},
			errExpected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			appTransport, err := gh.NewAppsTransport(test.githubClient.Transport, 123456, key)
			require.NoError(t, err)

			githubRepo := NewGithubRepo(appTransport, zaptest.NewLogger(t), test.githubClient)

			r, err := githubRepo.GetCommit(context.Background(),
				test.githubInstallID,
				test.githubRepoOwner,
				test.githubRepo,
				test.githubBranch)

			if test.errExpected != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResponse, r)
		})
	}
}

func TestGetRepo(t *testing.T) {
	tests := map[string]struct {
		githubInstallID  string
		githubRepoOwner  string
		githubRepoName   string
		githubClient     *http.Client
		expectedResponse *github.Repository
		errExpected      error
	}{
		"auth error": {
			githubInstallID: "1234567",
			githubRepoOwner: "octocat",
			githubRepoName:  "Hello-World",
			githubClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatchHandler(
					mock.PostAppInstallationsAccessTokensByInstallationId,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						mock.WriteError(
							w,
							http.StatusUnauthorized,
							"tokenError",
						)
					}),
				),
			),
			errExpected: errors.New("received non 2xx response status \"401 Unauthorized\""),
		},
		"happy path": {
			githubInstallID: "1234567",
			githubRepoOwner: "octocat",
			githubRepoName:  "Hello-World",
			githubClient: mock.NewMockedHTTPClient(
				mock.WithRequestMatch(
					mock.PostAppInstallationsAccessTokensByInstallationId,
					struct {
						Token     string    `json:"token"`
						ExpiresAt time.Time `json:"expires_at"`
					}{
						Token:     "this-is-the-token",
						ExpiresAt: time.Now().Add(5 * time.Minute),
					},
				),
				mock.WithRequestMatchHandler(
					mock.GetReposByOwnerByRepo,
					http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						_, err := w.Write(mock.MustMarshal(&github.Repository{
							ID:            toPtr(int64(123)),
							Name:          toPtr("Hello-World"),
							DefaultBranch: toPtr("main"),
							Private:       toPtr(false),
						}))
						assert.NoError(t, err)
					}),
				),
			),
			expectedResponse: &github.Repository{
				ID:            toPtr(int64(123)),
				Name:          toPtr("Hello-World"),
				DefaultBranch: toPtr("main"),
				Private:       toPtr(false),
			},
			errExpected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			appTransport, err := gh.NewAppsTransport(test.githubClient.Transport, 123456, key)
			require.NoError(t, err)

			githubRepo := NewGithubRepo(appTransport, zaptest.NewLogger(t), test.githubClient)

			r, err := githubRepo.GetRepo(context.Background(),
				test.githubInstallID,
				test.githubRepoOwner,
				test.githubRepoName)

			if test.errExpected != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, test.errExpected.Error())
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expectedResponse, r)
		})
	}
}
