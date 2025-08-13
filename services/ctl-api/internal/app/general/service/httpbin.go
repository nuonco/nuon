package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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
	QueryParams map[string][]string `json:"query_params"`
	IP          string              `json:"ip"`
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

	// Read request body
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
			QueryParams: ctx.Request.URL.Query(),
			IP:          ctx.ClientIP(),
		},
	}

	// Pretty print the entire response
	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		ctx.Error(fmt.Errorf("failed to marshal response: %w", err))
		return
	}

	wait := ctx.Query("delay")
	if wait != "" {
		delay, err := strconv.Atoi(wait)
		if err != nil {
			ctx.Error(fmt.Errorf("invalid delay: %w", err))
			return
		}
		time.Sleep(time.Duration(delay*1000) * time.Millisecond)
	}

	ctx.Header("Content-Type", "application/json")
	ctx.String(code, string(jsonBytes))
}
