package apps

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

var sha256ChecksumPattern = regexp.MustCompile(`(?i)^sha256:([0-9a-f]{64})$`)
var contentRangePattern = regexp.MustCompile(`^bytes ([0-9]+)-([0-9]+)/([0-9]+)$`)
var errChecksumMismatch = errors.New("download checksum mismatch")

type DownloadBundleOptions struct {
	File      string
	NoResume  bool
	Overwrite bool
}

func (s *Service) DownloadBundle(ctx context.Context, appID, bundleID string, opts DownloadBundleOptions) error {
	appID, err := s.bundleAppID(ctx, appID)
	if err != nil {
		return ui.PrintError(err)
	}
	grant, err := s.api.CreateReleasePackageDownloadGrant(ctx, bundleID)
	if err != nil {
		return ui.PrintError(err)
	}
	if grant == nil {
		return ui.PrintError(fmt.Errorf("download grant is empty"))
	}
	legacyGrant := &models.ServiceDownloadGrantResponse{
		URL: grant.URL, ExpiresAt: grant.ExpiresAt, Filename: grant.Filename, Size: grant.Size,
		TransportChecksum: grant.ArchiveChecksum, ManifestDigest: grant.ManifestDigest, SupportsRange: grant.SupportsRange,
	}
	if err := downloadBundle(ctx, http.DefaultClient, legacyGrant, opts); err != nil {
		return ui.PrintError(err)
	}
	ui.PrintSuccess(fmt.Sprintf("downloaded bundle to %s", opts.File))
	return nil
}

func downloadBundle(ctx context.Context, client *http.Client, grant *models.ServiceDownloadGrantResponse, opts DownloadBundleOptions) error {
	if grant == nil {
		return fmt.Errorf("download grant is empty")
	}
	checksumMatch := sha256ChecksumPattern.FindStringSubmatch(grant.TransportChecksum)
	if checksumMatch == nil {
		return fmt.Errorf("download grant has malformed transport checksum")
	}
	if grant.Size < 0 {
		return fmt.Errorf("download grant has invalid size")
	}
	if opts.File == "" {
		return fmt.Errorf("destination file path is required")
	}
	if _, err := os.Stat(opts.File); err == nil && !opts.Overwrite {
		return fmt.Errorf("destination already exists; use --overwrite to replace it")
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("inspect destination: %w", err)
	}

	partialPath := opts.File + ".partial"
	partialSize := int64(0)
	if info, err := os.Stat(partialPath); err == nil {
		partialSize = info.Size()
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect partial download: %w", err)
	}
	if opts.NoResume || !grant.SupportsRange || partialSize > grant.Size {
		partialSize = 0
	}
	if partialSize == grant.Size {
		err := verifyAndCommit(partialPath, opts.File, grant.Size, checksumMatch[1], opts.Overwrite)
		if errors.Is(err, errChecksumMismatch) && partialSize > 0 && !opts.NoResume {
			return downloadBundle(ctx, client, grant, DownloadBundleOptions{File: opts.File, NoResume: true, Overwrite: opts.Overwrite})
		}
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, grant.URL, nil)
	if err != nil {
		return fmt.Errorf("create download request: %w", err)
	}
	if partialSize > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", partialSize))
	}
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("download request failed: %w", ctx.Err())
		}
		return fmt.Errorf("download request failed")
	}
	defer resp.Body.Close()

	appendPartial := false
	switch {
	case partialSize > 0 && resp.StatusCode == http.StatusPartialContent:
		if err := validateContentRange(resp.Header.Get("Content-Range"), partialSize, grant.Size); err != nil {
			return fmt.Errorf("download returned an invalid content range")
		}
		appendPartial = true
	case resp.StatusCode == http.StatusOK:
		partialSize = 0
	case partialSize == 0 && resp.StatusCode == http.StatusPartialContent:
		return fmt.Errorf("download returned unexpected HTTP status %d", resp.StatusCode)
	default:
		return fmt.Errorf("download returned unexpected HTTP status %d", resp.StatusCode)
	}

	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if appendPartial {
		flags = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	}
	partial, err := os.OpenFile(partialPath, flags, 0o600)
	if err != nil {
		return fmt.Errorf("open partial download: %w", err)
	}
	remaining := grant.Size - partialSize
	readLimit := remaining
	if readLimit < int64(^uint64(0)>>1) {
		readLimit++
	}
	written, copyErr := io.Copy(partial, io.LimitReader(resp.Body, readLimit))
	syncErr := partial.Sync()
	closeErr := partial.Close()
	if copyErr != nil {
		return fmt.Errorf("download interrupted: %w", copyErr)
	}
	if syncErr != nil {
		return fmt.Errorf("sync partial download: %w", syncErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close partial download: %w", closeErr)
	}
	if written > remaining {
		return fmt.Errorf("download size mismatch: response exceeded expected size")
	}

	err = verifyAndCommit(partialPath, opts.File, grant.Size, checksumMatch[1], opts.Overwrite)
	if errors.Is(err, errChecksumMismatch) && partialSize > 0 && !opts.NoResume {
		return downloadBundle(ctx, client, grant, DownloadBundleOptions{File: opts.File, NoResume: true, Overwrite: opts.Overwrite})
	}
	return err
}

func validateContentRange(value string, expectedStart, expectedSize int64) error {
	parts := contentRangePattern.FindStringSubmatch(value)
	if parts == nil {
		return fmt.Errorf("invalid content range")
	}
	start, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return err
	}
	end, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return err
	}
	total, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return err
	}
	if start != expectedStart || end != expectedSize-1 || total != expectedSize {
		return fmt.Errorf("invalid content range")
	}
	return nil
}

func verifyAndCommit(partialPath, destination string, expectedSize int64, expectedChecksum string, overwrite bool) error {
	info, err := os.Stat(partialPath)
	if err != nil {
		return fmt.Errorf("inspect partial download: %w", err)
	}
	if info.Size() != expectedSize {
		return fmt.Errorf("download size mismatch: expected %d bytes, got %d", expectedSize, info.Size())
	}

	partial, err := os.Open(partialPath)
	if err != nil {
		return fmt.Errorf("open partial download for verification: %w", err)
	}
	hash := sha256.New()
	_, copyErr := io.Copy(hash, partial)
	closeErr := partial.Close()
	if copyErr != nil {
		return fmt.Errorf("verify partial download: %w", copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close partial download after verification: %w", closeErr)
	}
	actualChecksum := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actualChecksum, expectedChecksum) {
		return errChecksumMismatch
	}
	if overwrite {
		if err := os.Rename(partialPath, destination); err != nil {
			return fmt.Errorf("commit downloaded file: %w", err)
		}
		return nil
	}
	if err := os.Link(partialPath, destination); err != nil {
		return fmt.Errorf("commit downloaded file: %w", err)
	}
	if err := os.Remove(partialPath); err != nil {
		return fmt.Errorf("remove committed partial file: %w", err)
	}
	return nil
}
