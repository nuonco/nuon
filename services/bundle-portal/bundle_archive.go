package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
	"github.com/nuonco/nuon/pkg/runner/airgap/day2run"
	"github.com/nuonco/nuon/pkg/runner/oci/bundle"
)

type bundleActionDefinitions map[string]map[string]*day2.BundleActionDefinition

func loadBundleActionDefinitions(ctx context.Context, paths []string) (bundleActionDefinitions, error) {
	definitions := make(bundleActionDefinitions, len(paths))
	for _, archivePath := range paths {
		dir, err := os.MkdirTemp("", "nuon-bundle-portal-")
		if err != nil {
			return nil, err
		}
		archive, err := os.Open(archivePath)
		if err != nil {
			_ = os.RemoveAll(dir)
			return nil, fmt.Errorf("open bundle archive %s: %w", archivePath, err)
		}
		archiveDigest, extractErr := bundle.Extract(dir, archive)
		closeErr := archive.Close()
		if extractErr != nil {
			_ = os.RemoveAll(dir)
			return nil, fmt.Errorf("extract bundle archive %s: %w", archivePath, extractErr)
		}
		if closeErr != nil {
			_ = os.RemoveAll(dir)
			return nil, fmt.Errorf("close bundle archive %s: %w", archivePath, closeErr)
		}
		opened, err := bundle.Open(ctx, dir)
		if err != nil {
			_ = os.RemoveAll(dir)
			return nil, fmt.Errorf("read bundle archive %s: %w", archivePath, err)
		}
		info := day2run.BundleInfoFromManifest("", "", opened.Manifest, day2.BundleInfo{}.ActivatedAt)
		actions := make(map[string]*day2.BundleActionDefinition)
		for _, content := range info.Contents {
			if content.Kind == day2.BundleContentKindAction {
				actions[content.Name] = content.ActionDefinition
			}
		}
		definitions["sha256:"+archiveDigest] = actions
		if err := os.RemoveAll(dir); err != nil {
			return nil, fmt.Errorf("remove extracted bundle %s: %w", archivePath, err)
		}
	}
	return definitions, nil
}

func addHistoricalActionDefinitions(history []day2.BundleInfo, definitions bundleActionDefinitions) {
	for historyIndex := range history {
		actions := definitions[history[historyIndex].ArchiveDigest]
		for contentIndex := range history[historyIndex].Contents {
			content := &history[historyIndex].Contents[contentIndex]
			if content.Kind == day2.BundleContentKindAction && content.ActionDefinition == nil {
				content.ActionDefinition = actions[content.Name]
			}
		}
	}
}
