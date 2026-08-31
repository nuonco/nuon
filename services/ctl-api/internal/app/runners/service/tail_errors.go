package service

import (
	"context"
	"database/sql/driver"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func writeTailUnavailable(ctx *gin.Context) {
	ctx.Header("Retry-After", "1")
	ctx.JSON(http.StatusServiceUnavailable, stderr.ErrResponse{
		Error:       "service unavailable",
		Description: "the backing store did not respond before the long-poll deadline; retry the request",
	})
}

func tailProbeQueryError(ctx context.Context, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return errors.Join(err, ctxErr)
	}
	return err
}

func isTransientTailProbeError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, driver.ErrBadConn) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) && (netErr.Timeout() || netErr.Temporary()) {
		return true
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return strings.HasPrefix(pgErr.Code, "08") ||
		strings.HasPrefix(pgErr.Code, "53") ||
		pgErr.Code == "57P01" ||
		pgErr.Code == "57P02" ||
		pgErr.Code == "57P03"
}
