package service

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type HttpBinResponse struct {
	Code    int                `json:"code"`
	Text    string             `json:"text"`
	Request HttpBinRequestInfo `json:"request"`
}

type HttpBinRequestInfo struct {
	Method      string              `json:"method"`
	URI         string              `json:"uri"`
	Headers     map[string][]string `json:"headers"`
	Body        string              `json:"body"`
	Request     string              `json:"request"`
	QueryParams map[string][]string `json:"query_params"`
	IP          string              `json:"ip"`
	Panic       bool                `json:"panic"`
}

func (s *service) HttpBin(ctx *gin.Context) {
	code, err := strconv.Atoi(ctx.Param("code"))
	if err != nil {
		ctx.Error(fmt.Errorf("invalid status code: %w", err))
		return
	}

	text := http.StatusText(code)
	if text == "" {
		ctx.Error(fmt.Errorf("invalid status code: %d", code))
		return
	}

	panicArg := ctx.DefaultQuery("panic", "false")
	shouldPanic, err := strconv.ParseBool(panicArg)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to parse panic argument"))
		return
	}

	httpRequest, err := httputil.DumpRequest(ctx.Request, true)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to dump request"))
		return
	}
	httpRequestString := string(httpRequest)

	bodyBytes, _ := io.ReadAll(ctx.Request.Body)
	bodyString := string(bodyBytes)

	response := HttpBinResponse{
		Code: code,
		Text: text,
		Request: HttpBinRequestInfo{
			Method:      ctx.Request.Method,
			URI:         ctx.Request.URL.String(),
			Headers:     ctx.Request.Header,
			Body:        bodyString,
			Request:     httpRequestString,
			QueryParams: ctx.Request.URL.Query(),
			IP:          ctx.ClientIP(),
		},
	}

	wait := ctx.Query("delay")
	if wait != "" {
		delayDur, err := time.ParseDuration(wait)
		if err != nil {
			ctx.Error(fmt.Errorf("invalid delay: %w", err))
			return
		}
		time.Sleep(delayDur)
	}

	if shouldPanic {
		panic("HTTPBIN force panic")
	}

	ctx.IndentedJSON(code, response)
}
