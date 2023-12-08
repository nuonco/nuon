package activities

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type ListTestsRequest struct{}

type ListTestsResponse struct {
	Tests []string
}

func (a *Activities) ListTests(ctx context.Context, req *ListTestsRequest) (ListTestsResponse, error) {
	moduleDir := a.cfg.TestsDir
	resp := ListTestsResponse{
		Tests: make([]string, 0),
	}

	files, err := os.ReadDir(moduleDir)
	if err != nil {
		return resp, fmt.Errorf("unable to list tests: %w", err)
	}

	for _, fh := range files {
		if fh.IsDir() {
			continue
		}

		fp := filepath.Join(moduleDir, fh.Name())
		resp.Tests = append(resp.Tests, fp)
	}

	return resp, nil
}
