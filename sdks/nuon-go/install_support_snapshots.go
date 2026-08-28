package nuon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) CreateInstallSupportSnapshot(ctx context.Context, installID string, archive io.Reader) (*models.ServiceSupportSnapshotResponse, error) {
	req, err := c.newInstallSupportSnapshotRequest(ctx, http.MethodPost, installID, "", archive)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	var snapshot models.ServiceSupportSnapshotResponse
	if err := c.doInstallSupportSnapshotRequest(req, &snapshot, http.StatusOK, http.StatusCreated); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (c *client) ListInstallSupportSnapshots(ctx context.Context, installID string) ([]*models.ServiceSupportSnapshotResponse, error) {
	req, err := c.newInstallSupportSnapshotRequest(ctx, http.MethodGet, installID, "", nil)
	if err != nil {
		return nil, err
	}

	var snapshots []*models.ServiceSupportSnapshotResponse
	if err := c.doInstallSupportSnapshotRequest(req, &snapshots, http.StatusOK); err != nil {
		return nil, err
	}
	return snapshots, nil
}

func (c *client) GetInstallSupportSnapshot(ctx context.Context, installID, snapshotID string) (*models.ServiceSupportSnapshotResponse, error) {
	req, err := c.newInstallSupportSnapshotRequest(ctx, http.MethodGet, installID, snapshotID, nil)
	if err != nil {
		return nil, err
	}

	var snapshot models.ServiceSupportSnapshotResponse
	if err := c.doInstallSupportSnapshotRequest(req, &snapshot, http.StatusOK); err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (c *client) newInstallSupportSnapshotRequest(ctx context.Context, method, installID, snapshotID string, body io.Reader) (*http.Request, error) {
	path := fmt.Sprintf("%s/v1/installs/%s/support-snapshots", c.APIURL, url.PathEscape(installID))
	if snapshotID != "" {
		path += "/" + url.PathEscape(snapshotID)
	}
	req, err := http.NewRequestWithContext(ctx, method, path, body)
	if err != nil {
		return nil, fmt.Errorf("create install support snapshot request: %w", err)
	}
	return req, nil
}

func (c *client) doInstallSupportSnapshotRequest(req *http.Request, target any, expected ...int) error {
	resp, err := (&http.Client{Transport: c.appTransport}).Do(req)
	if err != nil {
		return fmt.Errorf("execute install support snapshot request: %w", err)
	}
	defer resp.Body.Close()

	for _, status := range expected {
		if resp.StatusCode == status {
			if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
				return fmt.Errorf("decode install support snapshot response: %w", err)
			}
			return nil
		}
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return fmt.Errorf("install support snapshot request returned status %d: %s", resp.StatusCode, string(body))
}
