package chaos

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"go.uber.org/zap"
)

func panicHandler(ctx *gin.Context, logger *zap.Logger) {
	logger.Error("chaos monkey panic triggered")
	panic("chaos monkey says no!")
}

func internalServerErrorHandler(ctx *gin.Context, logger *zap.Logger) {
	logger.Error("chaos monkey 500 error triggered")
	ctx.JSON(http.StatusInternalServerError, stderr.ErrSystem{
		Err:         fmt.Errorf("thanos snapped away half the shirts in the world for univeral shirts versus skins basketball game"),
		Description: "Internal server error triggered by a thanos snap",
	})
	ctx.Abort()
}

func serviceUnavailableHandler(ctx *gin.Context, logger *zap.Logger) {
	logger.Error("chaos monkey 503 service unavailable triggered")
	htmlResponse := `
<!DOCTYPE html>
<html>
<head>
    <title>Service Temporarily Unavailable</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; margin-top: 100px; }
        .error { color: #e74c3c; }
    </style>
</head>
<body>
    <h1 class="error">🐒 Service Temporarily Unavailable</h1>
    <p>Our thanos is having some fun. Please try again later!</p>
    <p><small>Error triggered by chaos engineering middleware</small></p>
</body>
</html>`
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(http.StatusServiceUnavailable, htmlResponse)
	ctx.Abort()
}

func badGatewayHandler(ctx *gin.Context, logger *zap.Logger) {
	htmlResponse := `
<!DOCTYPE html>
<html>
<head>
		<title>Bad Gateway</title>
		<style>
			body { font-family: Arial, sans-serif; text-align: center; margin-top: 100px; }
			.error { color: #e74c3c; }
		</style>
</head>	
<body>	
		<h1 class="error">🐵 Bad Gateway</h1>
		<p>Our thanos is causing some trouble. Please try again later!</p>
		<p><small>Error triggered by chaos engineering middleware</small></p>
</body>	
</html>`
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(http.StatusBadGateway, htmlResponse)
	logger.Error("thanos 502 bad gateway triggered")
	ctx.Abort()
}

func timeoutHandler(ctx *gin.Context, logger *zap.Logger) {
	logger.Error("thanos timeout triggered")
	ctx.JSON(http.StatusRequestTimeout, gin.H{
		"error": "Request timeout triggered by chaos monkey",
		"code":  "CHAOS_TIMEOUT",
	})
	ctx.Abort()
}
