package analytics

import (
	"github.com/powertoolsdev/mono/pkg/analytics/context"
	"go.uber.org/zap"
)

// ContextWriter provides the same meethods as Writer, but accepts a context instead of account and org information.
// It will read account and org data off of the provided context.
type ContextWriter struct {
	writer Writer
	logger *zap.Logger
}

// NewContextWriter creates a new ContextWriter.
func NewContextWriter(writeKey string, l *zap.Logger) *ContextWriter {
	return &ContextWriter{
		writer: NewWriter(writeKey),
		logger: l,
	}
}

// Identify reads the users's account from the context and identifies the user.
func (cw *ContextWriter) Identify(ctx context.Context) {
	account, err := context.AccountFromContext(ctx)
	if err != nil {
		cw.logger.Error(err.Error())
		return
	}
	cw.writer.Identify(account.ID, account.Email)
}

// Group reads the user's account and org from the context and groups the user.
func (cw *ContextWriter) Group(ctx context.Context) {
	account, err := context.AccountFromContext(ctx)
	if err != nil {
		cw.logger.Error(err.Error())
		return
	}
	org, err := context.OrgFromContext(ctx)
	if err != nil {
		cw.logger.Error(err.Error())
		return
	}
	cw.writer.Group(account.ID, org.ID, org.Name)
}

// Track reads the user's account from the context and tracks the user event.
func (cw *ContextWriter) Track(ctx context.Context, event Event, properties map[string]any) {
	account, err := context.AccountFromContext(ctx)
	if err != nil {
		cw.logger.Error(err.Error())
		return
	}
	cw.writer.Track(account.ID, event, properties)
}
