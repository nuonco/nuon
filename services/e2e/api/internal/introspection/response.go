package introspection

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OKResponse struct {
	Description string `json:"description"`
	Response    any    `json:"response"`
}

func (s *svc) writeOKResponse(ctx *gin.Context, resp OKResponse) {
	ctx.JSON(http.StatusOK, resp)
}

type ErrResponse struct {
	Description string `json:"description"`
	Err         error  `json:"-"`
	ErrString   string `json:"err"`
}

func (s *svc) writeErrResponse(ctx *gin.Context, resp ErrResponse) {
	l := zap.L()
	l.Error("recieved handler error", zap.Error(resp.Err))

	resp.ErrString = resp.Err.Error()
	ctx.JSON(http.StatusBadRequest, resp)
}
