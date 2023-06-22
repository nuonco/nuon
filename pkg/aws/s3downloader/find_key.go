package s3downloader

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

func (s *s3Downloader) FindKeys(ctx context.Context, match string, filename string) ([]string, error) {
	client, err := s.getClient(ctx)
	if err != nil {
		return nil, err
	}

	allKeys, err := s.listPrefix(ctx, client, "")
	if err != nil {
		return nil, fmt.Errorf("unable to list prefix: %w", err)
	}

	keys, err := s.findKeys(allKeys, match, filename)
	if err != nil {
		return nil, fmt.Errorf("unable to find keys: %w", err)
	}
	return keys, nil
}

func (s *s3Downloader) findKeys(allKeys []string, match string, filename string) ([]string, error) {
	keys := make([]string, 0)
	for _, key := range allKeys {
		matched, err := filepath.Match(match, key)
		if err != nil {
			return nil, fmt.Errorf("invalid match %s: %w", match, err)
		}
		if matched {
			keys = append(keys, key)
		}

		if !strings.Contains(key, match) {
			continue
		}

		if filename == "" {
			keys = append(keys, key)
		}

		if filepath.Base(key) == filename {
			keys = append(keys, key)
		}
	}

	return keys, nil
}
