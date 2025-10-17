package stderr

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/config"
)

// Response Writer that caches the response body
type CachedResponseWriter struct {
	gin.ResponseWriter
	Body *bytes.Buffer
}

func (w *CachedResponseWriter) Write(b []byte) (int, error) {
	w.Body.Write(b)
	return w.ResponseWriter.Write(b)
}

func NewCachedResponseWriter(w gin.ResponseWriter) *CachedResponseWriter {
	return &CachedResponseWriter{
		ResponseWriter: w,
		Body:           bytes.NewBufferString(""),
	}
}

type middleware struct {
	l *zap.Logger
}

func (m *middleware) Name() string {
	return "error"
}

func (m *middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Capture the request body
		var requestBody string
		if c.Request.Body != nil {
			b, _ := io.ReadAll(c.Request.Body)
			requestBody = string(b)
			c.Request.Body = io.NopCloser(bytes.NewReader(b)) // Restore the body for the actual handler to use
		}

		// Capture the response body
		writer := NewCachedResponseWriter(c.Writer)
		c.Writer = writer

		// Log server errors
		defer func() {
			m.LogErrors(c, requestBody, writer.Body.String())
		}()

		// Panic recovery (called last, executes first)
		// Execute this first because it writes the response after recovering
		defer func() {
			m.RecoverFromPanic(c)
		}()

		c.Next()

		if len(c.Errors) < 1 {
			return
		}

		err := c.Errors[0]

		// Check if this is a binding error
		if err.Type == gin.ErrorTypeBind {
			m.l.Error("response already set, this usually means the endpoint is using ctx.BindJSON instead of ctx.ShouldBindJSON")
			c.JSON(http.StatusBadRequest, ErrResponse{
				Error:       "invalid request format",
				UserError:   true,
				Description: err.Error(),
			})
			return
		}

		// define common error handlers here
		var uErr ErrUser
		if errors.As(err, &uErr) {
			c.JSON(http.StatusBadRequest, ErrResponse{
				Error:       err.Error(),
				UserError:   true,
				Description: uErr.Description,
			})
			return
		}

		var cfgErr config.ErrConfig
		if errors.As(err, &cfgErr) {
			c.JSON(http.StatusBadRequest, ErrResponse{
				Error:       err.Error(),
				UserError:   true,
				Description: cfgErr.Description,
			})
			return
		}

		var authnErr ErrAuthentication
		if errors.As(err, &authnErr) {
			c.JSON(http.StatusUnauthorized, ErrResponse{
				Error:       err.Error(),
				UserError:   true,
				Description: authnErr.Description,
			})
			return
		}

		var sysErr ErrSystem
		if errors.As(err, &sysErr) {
			c.JSON(http.StatusInternalServerError, ErrResponse{
				Error:       err.Error(),
				UserError:   false,
				Description: sysErr.Description,
			})
			return
		}

		var nrErr ErrNotReady
		if errors.As(err, &nrErr) {
			// NOTE(jm): there really is not a good status code for "not ready".
			//
			// our options are:
			// 503 which implies a service issue.
			// 404 which implies not found
			// 3xx
			c.JSON(http.StatusConflict, ErrResponse{
				Error:       err.Error(),
				UserError:   true,
				Description: nrErr.Description,
			})
			return
		}

		var authzErr ErrAuthorization
		if errors.As(err, &authzErr) {
			c.JSON(http.StatusForbidden, ErrResponse{
				Error:       err.Error(),
				UserError:   true,
				Description: authzErr.Description,
			})
			return
		}

		// gorm not found errors are usually user errors
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrResponse{
				Error:       err.Error(),
				UserError:   true,
				Description: "not found",
			})
			return
		}

		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusBadRequest, ErrResponse{
				Error:       err.Error(),
				UserError:   true,
				Description: "duplicate key",
			})
			return
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "25303" || pgErr.Code == "23503" {
				c.JSON(http.StatusBadRequest, ErrResponse{
					Error:       err.Error(),
					UserError:   true,
					Description: "invalid foreign key - usually from using an invalid parent object ID",
				})
				return
			}
		}

		// validation errors for any request inputs
		var vErr validator.ValidationErrors
		if errors.As(err, &vErr) {
			c.JSON(http.StatusBadRequest, ErrResponse{
				Error:       fmt.Sprintf("invalid input for %s", vErr[0].Field()),
				UserError:   true,
				Description: fmt.Sprintf("invalid request input: %s", err),
			})
			return
		}

		// bad or unparseable request
		var ivReqErr ErrInvalidRequest
		if errors.As(err, &ivReqErr) {
			c.JSON(http.StatusBadRequest, ErrResponse{
				Error:       "invalid request",
				UserError:   true,
				Description: fmt.Sprintf("invalid request input: %s", err),
			})
			return
		}

		if errors.Is(err, context.DeadlineExceeded) {
			c.JSON(http.StatusInternalServerError, ErrResponse{
				Error:       "timeout",
				UserError:   true,
				Description: "we were unable to complete this request within time.",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrResponse{
			Error:       err.Error(),
			UserError:   false,
			Description: err.Error(),
		})
	}
}

func (m *middleware) RecoverFromPanic(c *gin.Context) {
	if r := recover(); r != nil {
		// Log the panic
		m.l.Error("panic recovered",
			zap.Any("panic", r),
			zap.Stack("stack"),
		)

		// Return a system error response
		c.JSON(http.StatusInternalServerError, ErrResponse{
			Error:       "internal server error",
			UserError:   false,
			Description: "An unexpected error occurred",
		})
		c.Abort()
	}
}

func (m *middleware) LogErrors(c *gin.Context, requestBody, responseBody string) {
	// Log errors for status >= 500
	if c.Writer.Status() >= 500 {
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.String("request_body", requestBody),
			zap.String("response_body", responseBody),
			zap.Stack("stack"),
		}

		var msg string
		if len(c.Errors) > 0 {
			errorList := strings.Join(c.Errors.Errors(), ", ")
			msg = fmt.Sprintf("internal server error. errors: %s", errorList)
		} else {
			msg = "internal server error."
		}
		m.l.Error(msg, fields...)
	}
}

func New(l *zap.Logger) *middleware {
	return &middleware{
		l: l,
	}
}
