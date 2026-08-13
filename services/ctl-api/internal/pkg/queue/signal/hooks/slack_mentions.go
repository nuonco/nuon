package hooks

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
)

// slackUserResolver resolves emails to workspace user ids via
// users.lookupByEmail, memoised per publish() call. User ids are
// workspace-scoped so the cache is keyed by (team, email). Misses (including
// users_not_found for people outside the workspace and missing_scope for
// installations that predate the users:read.email grant) are cached as ""
// so a publish never repeats a failing lookup.
type slackUserResolver struct {
	client *slackclient.Client
	l      *zap.Logger
	cache  map[string]string
}

func newSlackUserResolver(client *slackclient.Client, l *zap.Logger) *slackUserResolver {
	if l == nil {
		l = zap.NewNop()
	}
	return &slackUserResolver{client: client, l: l, cache: map[string]string{}}
}

// resolve returns the workspace user id for email, or "" when the email
// can't be resolved. Fail-open by design: lookup errors are debug-logged
// and never propagate — the renderer falls back to the plain email.
func (r *slackUserResolver) resolve(ctx context.Context, botToken, teamID, email string) string {
	if r == nil || r.client == nil || email == "" || !strings.Contains(email, "@") {
		return ""
	}

	key := teamID + "|" + email
	if id, ok := r.cache[key]; ok {
		return id
	}

	resp, err := r.client.UsersLookupByEmail(ctx, botToken, email)
	if err != nil {
		r.cache[key] = ""
		r.l.Debug("slack user lookup by email failed",
			zap.String("team_id", teamID),
			zap.Error(err))
		return ""
	}

	r.cache[key] = resp.User.ID
	return resp.User.ID
}

// withResolvedMentions returns a copy of rendered with the event's person
// emails (workflow creator, approval responder) resolved to workspace user
// ids for install's team. The shared rendered value is never mutated — it
// is reused across teams and user ids differ per workspace.
func (h *SlackSignalLifecycleHook) withResolvedMentions(
	ctx context.Context,
	resolver *slackUserResolver,
	install *app.SlackInstallation,
	rendered renderEvent,
) renderEvent {
	candidates := []string{rendered.event.Workflow.CreatedByEmail}
	if rendered.event.Approval != nil {
		candidates = append(candidates, strings.TrimSpace(rendered.event.Approval.RespondedBy))
	}

	var ids map[string]string
	for _, email := range candidates {
		if email == "" {
			continue
		}
		id := resolver.resolve(ctx, install.BotAccessToken, install.TeamID, email)
		if id == "" {
			continue
		}
		if ids == nil {
			ids = map[string]string{}
		}
		ids[email] = id
	}

	out := rendered
	out.event.SlackUserIDByEmail = ids
	return out
}
