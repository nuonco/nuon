package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/go-github/v50/github"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/worker/activities"
)

type webhookGH struct {
	hookID    int64
	err       error
	failCalls int
	calls     int
	urls      []string

	knownHooks map[string]int64
	findErr    error
	findCalls  int
}

func (s *webhookGH) GetInstallation(context.Context, string) (*github.Installation, error) {
	return nil, nil
}

func (s *webhookGH) ListInstallationRepos(context.Context, *app.VCSConnection) ([]*github.Repository, error) {
	return nil, nil
}

func (s *webhookGH) CreateOrgWebhook(_ context.Context, _ *app.VCSConnection, webhookURL, _ string) (int64, error) {
	s.calls++
	s.urls = append(s.urls, webhookURL)
	if s.err != nil && s.calls <= s.failCalls {
		return 0, s.err
	}
	return s.hookID, nil
}

func (s *webhookGH) FindOrgWebhookByURL(_ context.Context, _ *app.VCSConnection, webhookURL string) (int64, error) {
	s.findCalls++
	if s.findErr != nil {
		return 0, s.findErr
	}
	return s.knownHooks[webhookURL], nil
}

func (s *VCSServiceTestSuite) webhookActivities(gh activities.GithubClient) *activities.Activities {
	return activities.New(activities.Params{
		L:        s.service.L,
		Cfg:      &internal.Config{PublicAPIURL: "https://api.example.com"},
		DB:       s.service.DB,
		GhClient: gh,
	})
}

func (s *VCSServiceTestSuite) newWebhookConnection(name string) *app.VCSConnection {
	installID := fmt.Sprintf("%s-%d", name, time.Now().UnixNano())
	conn := &app.VCSConnection{
		OrgID:             s.testOrg.ID,
		CreatedByID:       s.testAcc.ID,
		GithubInstallID:   installID,
		GithubAccountID:   installID,
		GithubAccountName: "acme",
	}
	require.NoError(s.T(), s.service.DB.WithContext(s.ctx).Create(conn).Error)
	return conn
}

func (s *VCSServiceTestSuite) TestCreateWebhookSubscription_RetryAfterGitHubFailureUsesPersistedRow() {
	conn := s.newWebhookConnection("gh-retry-failure")
	gh := &webhookGH{hookID: 99, err: errors.New("github down"), failCalls: 1}
	acts := s.webhookActivities(gh)

	_, err := acts.CreateWebhookSubscription(s.ctx, activities.CreateWebhookSubscriptionRequest{
		VCSConnectionID: conn.ID,
	})
	require.EqualError(s.T(), err, "unable to create github org webhook: github down")

	var leftover app.VCSWebhookSubscription
	err = s.service.DB.Unscoped().
		Where(app.VCSWebhookSubscription{GithubInstallID: conn.GithubInstallID}).
		First(&leftover).Error
	require.ErrorIs(s.T(), err, gorm.ErrRecordNotFound)

	resp, err := acts.CreateWebhookSubscription(s.ctx, activities.CreateWebhookSubscriptionRequest{
		VCSConnectionID: conn.ID,
	})
	require.NoError(s.T(), err)
	require.False(s.T(), resp.AlreadyExisted)

	var saved app.VCSWebhookSubscription
	require.NoError(s.T(), s.service.DB.
		Where(app.VCSWebhookSubscription{ID: resp.SubscriptionID}).
		First(&saved).Error)
	require.Equal(s.T(), int64(99), saved.GithubHookID)
	require.Equal(s.T(), fmt.Sprintf("https://api.example.com/v1/vcs/webhooks/%s/events", saved.ID), resp.WebhookURL)
	require.Equal(s.T(), resp.WebhookURL, gh.urls[len(gh.urls)-1])
}

func (s *VCSServiceTestSuite) TestCreateWebhookSubscription_SoftDeletedRowDoesNotClaimSuccess() {
	conn := s.newWebhookConnection("gh-retry-deleted")
	stale := &app.VCSWebhookSubscription{
		OrgID:           conn.OrgID,
		CreatedByID:     s.testAcc.ID,
		VCSConnectionID: conn.ID,
		GithubInstallID: conn.GithubInstallID,
		WebhookSecret:   "stale-secret",
	}
	require.NoError(s.T(), s.service.DB.WithContext(s.ctx).Create(stale).Error)
	require.NoError(s.T(), s.service.DB.WithContext(s.ctx).Delete(stale).Error)

	gh := &webhookGH{hookID: 42}
	resp, err := s.webhookActivities(gh).CreateWebhookSubscription(s.ctx, activities.CreateWebhookSubscriptionRequest{
		VCSConnectionID: conn.ID,
	})
	require.NoError(s.T(), err)
	require.NotEqual(s.T(), stale.ID, resp.SubscriptionID)

	var saved app.VCSWebhookSubscription
	require.NoError(s.T(), s.service.DB.
		Where(app.VCSWebhookSubscription{ID: resp.SubscriptionID}).
		First(&saved).Error)
	require.Equal(s.T(), int64(42), saved.GithubHookID)
	require.Contains(s.T(), resp.WebhookURL, saved.ID)
	require.NotContains(s.T(), gh.urls[0], stale.ID)
}

func (s *VCSServiceTestSuite) newUnfinishedSubscription(conn *app.VCSConnection) *app.VCSWebhookSubscription {
	sub := &app.VCSWebhookSubscription{
		OrgID:           conn.OrgID,
		CreatedByID:     s.testAcc.ID,
		VCSConnectionID: conn.ID,
		GithubInstallID: conn.GithubInstallID,
		WebhookSecret:   "pending-secret",
	}
	require.NoError(s.T(), s.service.DB.WithContext(s.ctx).Create(sub).Error)
	return sub
}

func (s *VCSServiceTestSuite) TestCreateWebhookSubscription_UnfinishedRowIsCompleted() {
	conn := s.newWebhookConnection("gh-retry-unfinished")
	pending := s.newUnfinishedSubscription(conn)

	gh := &webhookGH{hookID: 55}
	resp, err := s.webhookActivities(gh).CreateWebhookSubscription(s.ctx, activities.CreateWebhookSubscriptionRequest{
		VCSConnectionID: conn.ID,
	})
	require.NoError(s.T(), err)
	require.False(s.T(), resp.AlreadyExisted)
	require.Equal(s.T(), pending.ID, resp.SubscriptionID)

	var saved app.VCSWebhookSubscription
	require.NoError(s.T(), s.service.DB.
		Where(app.VCSWebhookSubscription{ID: pending.ID}).
		First(&saved).Error)
	require.Equal(s.T(), int64(55), saved.GithubHookID)
	require.Equal(s.T(), resp.WebhookURL, saved.WebhookURL)
	require.Equal(s.T(), 1, gh.calls)
}

func (s *VCSServiceTestSuite) TestCreateWebhookSubscription_UnfinishedRowAdoptsExistingHook() {
	conn := s.newWebhookConnection("gh-retry-adopt")
	pending := s.newUnfinishedSubscription(conn)
	webhookURL := fmt.Sprintf("https://api.example.com/v1/vcs/webhooks/%s/events", pending.ID)

	gh := &webhookGH{
		err:        errors.New("422 Hook already exists on this organization"),
		failCalls:  10,
		knownHooks: map[string]int64{webhookURL: 77},
	}
	acts := s.webhookActivities(gh)
	resp, err := acts.CreateWebhookSubscription(s.ctx, activities.CreateWebhookSubscriptionRequest{
		VCSConnectionID: conn.ID,
	})
	require.NoError(s.T(), err)
	require.Equal(s.T(), pending.ID, resp.SubscriptionID)
	require.Equal(s.T(), webhookURL, resp.WebhookURL)
	require.Equal(s.T(), 1, gh.findCalls)

	var saved app.VCSWebhookSubscription
	require.NoError(s.T(), s.service.DB.
		Where(app.VCSWebhookSubscription{ID: pending.ID}).
		First(&saved).Error)
	require.Equal(s.T(), int64(77), saved.GithubHookID)
	require.Equal(s.T(), webhookURL, saved.WebhookURL)

	resp, err = acts.CreateWebhookSubscription(s.ctx, activities.CreateWebhookSubscriptionRequest{
		VCSConnectionID: conn.ID,
	})
	require.NoError(s.T(), err)
	require.True(s.T(), resp.AlreadyExisted)
	require.Equal(s.T(), 1, gh.calls)
}

func (s *VCSServiceTestSuite) TestCreateWebhookSubscription_UnfinishedRowSurvivesGitHubOutage() {
	conn := s.newWebhookConnection("gh-retry-outage")
	pending := s.newUnfinishedSubscription(conn)

	gh := &webhookGH{err: errors.New("github down"), failCalls: 10, findErr: errors.New("github down")}
	_, err := s.webhookActivities(gh).CreateWebhookSubscription(s.ctx, activities.CreateWebhookSubscriptionRequest{
		VCSConnectionID: conn.ID,
	})
	require.EqualError(s.T(), err, "unable to create github org webhook: github down")

	var saved app.VCSWebhookSubscription
	require.NoError(s.T(), s.service.DB.Unscoped().
		Where(app.VCSWebhookSubscription{ID: pending.ID}).
		First(&saved).Error)
	require.Zero(s.T(), saved.DeletedAt)
	require.Zero(s.T(), saved.GithubHookID)
}

func (s *VCSServiceTestSuite) TestCreateWebhookSubscription_SoftDeletedRowWithHookIsKept() {
	conn := s.newWebhookConnection("gh-retry-deleted-live")
	retired := &app.VCSWebhookSubscription{
		OrgID:           conn.OrgID,
		CreatedByID:     s.testAcc.ID,
		VCSConnectionID: conn.ID,
		GithubInstallID: conn.GithubInstallID,
		WebhookSecret:   "secret",
		WebhookURL:      "https://api.example.com/v1/vcs/webhooks/retired/events",
		GithubHookID:    9,
	}
	require.NoError(s.T(), s.service.DB.WithContext(s.ctx).Create(retired).Error)
	require.NoError(s.T(), s.service.DB.WithContext(s.ctx).Delete(retired).Error)

	gh := &webhookGH{hookID: 100}
	_, err := s.webhookActivities(gh).CreateWebhookSubscription(s.ctx, activities.CreateWebhookSubscriptionRequest{
		VCSConnectionID: conn.ID,
	})
	require.Error(s.T(), err)
	require.Zero(s.T(), gh.calls)

	var kept app.VCSWebhookSubscription
	require.NoError(s.T(), s.service.DB.Unscoped().
		Where(app.VCSWebhookSubscription{ID: retired.ID}).
		First(&kept).Error)
	require.NotZero(s.T(), kept.DeletedAt)
}

func (s *VCSServiceTestSuite) TestCreateWebhookSubscription_ExistingHookIsLeftAlone() {
	conn := s.newWebhookConnection("gh-retry-exists")
	existing := &app.VCSWebhookSubscription{
		OrgID:           conn.OrgID,
		CreatedByID:     s.testAcc.ID,
		VCSConnectionID: conn.ID,
		GithubInstallID: conn.GithubInstallID,
		WebhookSecret:   "secret",
		WebhookURL:      "https://api.example.com/v1/vcs/webhooks/already/events",
		GithubHookID:    7,
	}
	require.NoError(s.T(), s.service.DB.WithContext(s.ctx).Create(existing).Error)

	gh := &webhookGH{hookID: 100}
	resp, err := s.webhookActivities(gh).CreateWebhookSubscription(s.ctx, activities.CreateWebhookSubscriptionRequest{
		VCSConnectionID: conn.ID,
	})
	require.NoError(s.T(), err)
	require.True(s.T(), resp.AlreadyExisted)
	require.Equal(s.T(), existing.ID, resp.SubscriptionID)
	require.Equal(s.T(), existing.WebhookURL, resp.WebhookURL)
	require.Zero(s.T(), gh.calls)
}
