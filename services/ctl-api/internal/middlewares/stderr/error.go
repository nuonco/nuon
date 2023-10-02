package stderr

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ErrUser struct {
	Err         error
	Description string
}

func (u ErrUser) Error() string {
	return u.Err.Error()
}

type ErrResponse struct {
	Error       string `json:"error"`
	UserError   bool   `json:"user_error"`
	Description string `json:"description"`
}

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

		// validation errors for any request inputs
		var vErr validator.ValidationErrors
		if errors.As(err, &vErr) {
			c.JSON(http.StatusBadRequest, ErrResponse{
				Error:       err.Error(),
				UserError:   true,
				Description: "invalid request input",
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
