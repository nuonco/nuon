package stderr

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type middleware struct {
	l *zap.Logger
}

func (m *middleware) Name() string {
	return "error"
}

func (m *middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) < 1 {
			return
		}

		err := c.Errors[0]

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

		var authnErr ErrAuthentication
		if errors.As(err, &authnErr) {
			c.JSON(http.StatusUnauthorized, ErrResponse{
				Error:       err.Error(),
				UserError:   true,
				Description: authnErr.Description,
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

		c.JSON(http.StatusInternalServerError, ErrResponse{
			Error:       err.Error(),
			UserError:   true,
			Description: err.Error(),
		})
	}
}

func New(l *zap.Logger) *middleware {
	return &middleware{
		l: l,
	}
}
