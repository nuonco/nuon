package apps

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type bundleDownloadIdentity struct {
	AppID    string `json:"app_id"`
	BundleID string `json:"bundle_id"`
	Checksum string `json:"transport_checksum"`
	Size     int64  `json:"size"`
}

type bundleDownloadResult struct {
	bundleDownloadIdentity
	File     string `json:"file"`
	Verified bool   `json:"verified"`
}

func (s *Service) DownloadBundle(ctx context.Context, appID, bundleID, destination string, force, asJSON bool) error {
	appID, err := s.bundleAppID(ctx, appID)
	if err != nil {
		return ui.PrintError(err)
	}
	bundle, err := s.api.GetAppBundle(ctx, appID, bundleID)
	if err != nil {
		return ui.PrintError(err)
	}
	if bundle.Status != "active" || bundle.VerifiedAt == "" {
		return ui.PrintError(&ui.CLIUserError{Msg: fmt.Sprintf("bundle %s is not ready (status: %s). Wait: nuon apps bundles wait --app-id %s --bundle-id %s", bundleID, bundle.Status, appID, bundleID)})
	}
	spinner := bubbles.NewSpinnerView(asJSON, s.cfg.Interactive)
	spinner.Start("Downloading bundle")
	progress := func(written, total int64) {
		spinner.Update(fmt.Sprintf("Downloading %s: %.0f%% (%d / %d bytes)", bundleID, float64(written)/float64(total)*100, written, total))
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment, ForceAttemptHTTP2: true,
			DialContext:         (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second, ResponseHeaderTimeout: 30 * time.Second,
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Scheme != "https" || len(via) >= 10 {
				return errors.New("unsafe or excessive storage redirects")
			}
			return nil
		},
	}
	defer client.CloseIdleConnections()
	result, err := downloadBundleArchive(ctx, s.api, client, bundle, destination, force, progress)
	if err != nil {
		if !asJSON {
			spinner.Fail(err)
		}
		return ui.PrintError(&ui.CLIUserError{Msg: err.Error()})
	}
	if asJSON {
		ui.PrintJSON(result)
	} else {
		spinner.Success("SHA-256 verified")
		ui.Printf("Saved: %s (%d bytes)\n", result.File, result.Size)
	}
	return nil
}

func downloadBundleArchive(ctx context.Context, api nuon.Client, client *http.Client, bundle *models.ServiceBundleResponse, destination string, force bool, progress func(int64, int64)) (*bundleDownloadResult, error) {
	checksum := strings.TrimPrefix(strings.ToLower(bundle.TransportChecksum), "sha256:")
	digest, err := hex.DecodeString(checksum)
	if err != nil || len(digest) != sha256.Size || bundle.Size <= 0 {
		return nil, errors.New("bundle has invalid archive size or SHA-256 metadata")
	}
	grant, err := api.CreateAppBundleDownloadGrant(ctx, bundle.AppID, bundle.ID)
	if err != nil {
		return nil, err
	}
	if destination == "" {
		destination = grant.Filename
		if destination == "" || destination == "." || filepath.Base(destination) != destination || strings.ContainsAny(destination, "/\\") {
			return nil, errors.New("download grant has an unsafe filename; provide --file")
		}
	}
	destination, err = filepath.Abs(destination)
	if err != nil {
		return nil, err
	}
	if info, err := os.Lstat(destination); err == nil {
		if !info.Mode().IsRegular() || !force {
			return nil, fmt.Errorf("destination %s already exists; use --force to replace a regular file", destination)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	identity := bundleDownloadIdentity{AppID: bundle.AppID, BundleID: bundle.ID, Checksum: "sha256:" + checksum, Size: bundle.Size}
	partial, metadata := destination+".part", destination+".part.json"
	file, err := openBundlePartial(partial, metadata, identity)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	for attempt := 0; attempt < 3; attempt++ {
		if grant.Size != identity.Size || strings.TrimPrefix(strings.ToLower(grant.TransportChecksum), "sha256:") != checksum {
			return nil, errors.New("download grant does not match the pinned bundle size and checksum")
		}
		offset, err := file.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, err
		}
		if offset > identity.Size {
			return nil, fmt.Errorf("partial download exceeds expected size; remove %s and %s and retry", partial, metadata)
		}
		progress(offset, identity.Size)
		if offset == identity.Size {
			break
		}
		if offset > 0 && !grant.SupportsRange {
			return nil, errors.New("storage does not support resuming this partial download")
		}
		u, err := url.Parse(grant.URL)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
			return nil, errors.New("download grant must contain an absolute HTTPS URL")
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, grant.URL, nil)
		if err != nil {
			return nil, errors.New("invalid download request")
		}
		req.Header.Set("Accept-Encoding", "identity")
		if offset > 0 {
			req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
		}
		resp, err := client.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, errors.New("storage download request failed; rerun the same command to resume")
		}
		if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
			resp.Body.Close()
			if attempt == 2 {
				return nil, fmt.Errorf("storage rejected the download after refreshing grants (HTTP %d)", resp.StatusCode)
			}
			grant, err = api.CreateAppBundleDownloadGrant(ctx, bundle.AppID, bundle.ID)
			if err != nil {
				return nil, err
			}
			continue
		}
		if resp.StatusCode == http.StatusOK {
			offset = 0
			if err := file.Truncate(0); err != nil {
				resp.Body.Close()
				return nil, err
			}
			if _, err := file.Seek(0, io.SeekStart); err != nil {
				resp.Body.Close()
				return nil, err
			}
		} else if resp.StatusCode != http.StatusPartialContent || resp.Header.Get("Content-Range") != fmt.Sprintf("bytes %d-%d/%d", offset, identity.Size-1, identity.Size) {
			resp.Body.Close()
			return nil, fmt.Errorf("storage returned an unexpected status or range (HTTP %d)", resp.StatusCode)
		}
		remaining := identity.Size - offset
		if resp.ContentLength >= 0 && resp.ContentLength != remaining {
			resp.Body.Close()
			return nil, errors.New("storage response length does not match the pinned bundle size")
		}
		writer := &bundleProgressWriter{writer: file, written: offset, total: identity.Size, progress: progress}
		copied, copyErr := io.Copy(writer, io.LimitReader(resp.Body, remaining+1))
		resp.Body.Close()
		if copyErr != nil {
			return nil, fmt.Errorf("download interrupted; rerun the same command to resume from %s", partial)
		}
		if copied != remaining {
			return nil, errors.New("downloaded byte count does not match the pinned bundle size; rerun to resume")
		}
		progress(identity.Size, identity.Size)
		break
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	if err != nil {
		return nil, err
	}
	if size != identity.Size || hex.EncodeToString(hash.Sum(nil)) != checksum {
		file.Close()
		os.Remove(partial)
		os.Remove(metadata)
		return nil, errors.New("bundle SHA-256 verification failed; corrupt partial download removed")
	}
	if err := file.Sync(); err != nil {
		return nil, err
	}
	if err := file.Close(); err != nil {
		return nil, err
	}
	if force {
		err = os.Rename(partial, destination)
	} else {
		err = os.Link(partial, destination)
		if err == nil {
			err = os.Remove(partial)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("unable to finalize verified bundle: %w", err)
	}
	if err := os.Remove(metadata); err != nil {
		return nil, err
	}
	return &bundleDownloadResult{bundleDownloadIdentity: identity, File: destination, Verified: true}, nil
}

func openBundlePartial(partial, metadata string, identity bundleDownloadIdentity) (*os.File, error) {
	info, err := os.Lstat(partial)
	if err == nil {
		if !info.Mode().IsRegular() {
			return nil, errors.New("partial download must be a regular file")
		}
		data, err := os.ReadFile(metadata)
		if err != nil {
			return nil, fmt.Errorf("partial download is missing its identity metadata; remove %s and retry", partial)
		}
		var stored bundleDownloadIdentity
		if err := json.Unmarshal(data, &stored); err != nil || stored != identity {
			return nil, errors.New("partial download belongs to a different bundle; use a different --file path or remove the partial files")
		}
		return os.OpenFile(partial, os.O_RDWR, 0600)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	data, err := json.Marshal(identity)
	if err != nil {
		return nil, err
	}
	meta, err := os.OpenFile(metadata, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, fmt.Errorf("unable to create partial metadata; remove stale %s and retry: %w", metadata, err)
	}
	_, writeErr := meta.Write(data)
	closeErr := meta.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		os.Remove(metadata)
		return nil, err
	}
	file, err := os.OpenFile(partial, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		os.Remove(metadata)
	}
	return file, err
}

type bundleProgressWriter struct {
	writer   io.Writer
	written  int64
	total    int64
	last     time.Time
	progress func(int64, int64)
}

func (w *bundleProgressWriter) Write(data []byte) (int, error) {
	n, err := w.writer.Write(data)
	w.written += int64(n)
	if time.Since(w.last) >= 500*time.Millisecond {
		w.progress(w.written, w.total)
		w.last = time.Now()
	}
	return n, err
}
