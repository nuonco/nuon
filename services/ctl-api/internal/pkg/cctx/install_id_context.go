package cctx

import (
	"context"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

// installIDSlot is stored by pointer so SetInstallIDContext updates the context
// the MCP middleware already passed into the tool handler.
type installIDSlot struct {
	id string
}

func SetInstallIDContext(ctx context.Context, installID string) context.Context {
	if ctx == nil {
		return nil
	}
	installID = strings.TrimSpace(installID)
	if slot, ok := ctx.Value(keys.InstallIDCtxKey).(*installIDSlot); ok {
		if installID != "" {
			slot.id = installID
		}
		return ctx
	}
	return context.WithValue(ctx, keys.InstallIDCtxKey, &installIDSlot{id: installID})
}

func InstallIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	slot, _ := ctx.Value(keys.InstallIDCtxKey).(*installIDSlot)
	if slot == nil {
		return ""
	}
	return slot.id
}
