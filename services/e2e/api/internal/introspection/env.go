package introspection

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

const envDescription = "Returns the entire environment of the running service."

func (s *svc) GetEnvHandler(ctx *gin.Context) {
	resp, err := s.getEnvHandler(ctx)
	if err != nil {
		s.writeErrResponse(ctx, ErrResponse{
			Description: envDescription,
			Err:         err,
		})
		return
	}

	s.writeOKResponse(ctx, OKResponse{
		Description: envDescription,
		Response:    resp,
	})
}

func (s *svc) getEnvHandler(ctx *gin.Context) (map[string]string, error) {
	env := make(map[string]string)
	for _, envStr := range os.Environ() {
		pieces := strings.SplitN(envStr, "=", 2)
		if len(pieces) != 2 {
			return nil, fmt.Errorf("invalid environment var: %s", envStr)
		}
		env[pieces[0]] = pieces[1]
	}

	return env, nil
}

func (s *svc) getEnvByPrefix(prefix string) (map[string]string, error) {
	resp := make(map[string]string, 0)

	for _, envStr := range os.Environ() {
		pieces := strings.SplitN(envStr, "=", 2)
		if len(pieces) != 2 {
			return nil, fmt.Errorf("invalid environment variable: %s", envStr)
		}

		k, v := pieces[0], pieces[1]
		if !strings.HasPrefix(k, prefix) {
			continue
		}

		k = strings.TrimPrefix(k, prefix)
		k = strings.TrimPrefix(k, "_")
		k = strings.ToLower(k)
		resp[k] = v
	}

	return resp, nil
}
