package service

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *service) logFlowAPIAction(ctx *gin.Context, flowEvent string, fields ...zap.Field) {
	l := cctx.GetLogger(ctx, s.l)

	base := make([]zap.Field, 0, len(fields)+2)
	base = append(base, zap.String("flow_event", flowEvent))
	if acct, err := cctx.AccountFromGinContext(ctx); err == nil && acct != nil {
		base = append(base, zap.String("acting_account_email", acct.Email))
	}

	l.Info("flow telemetry", append(base, fields...)...)
}
