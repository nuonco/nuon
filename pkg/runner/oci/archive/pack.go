package ociarchive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	orasfile "oras.land/oras-go/v2/content/file"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/op"
)

const (
	defaultArtifactType string = "artifact/nuon"
	defaultLocalTag     string = "latest"
	tarLayerMediaType   string = "application/vnd.nuon.archive.layer.v1.tar+gzip"
	tarLayerTitle       string = "."
	maxArchiveLayers           = 400
)

type FileRef struct {
	AbsPath  string `mapstructure:"abs_path,omitempty"`
	RelPath  string `mapstructure:"rel_path,omitempty"`
	FileType string `mapstructure:"file_type,omitempty"`
}

type archiveLayer struct {
	title string
	files []FileRef
}

func (r *archive) Pack(ctx context.Context, log *zap.Logger, filePaths []FileRef) (retErr error) {
	opCtx, end := op.Tool(ctx, "oci", "pack")
	ctx = opCtx
	defer func() { end(retErr) }()
	if l, err := pkgctx.Logger(ctx); err == nil && l != nil {
		log = l
	}

	nonEmptyFiles := make([]FileRef, 0, len(filePaths))
	for _, file := range filePaths {
		stat, err := os.Stat(file.AbsPath)
		if err != nil {
			return fmt.Errorf("unable to stat file: %w", err)
		}
		if stat.Size() < 1 {
			log.Info("skipping empty file", zap.String("path", file.RelPath))
			continue
		}
		nonEmptyFiles = append(nonEmptyFiles, file)
	}

	archiveLayers := groupArchiveLayers(nonEmptyFiles)
	layers := make([]v1.Descriptor, 0, len(archiveLayers))
	for i, archiveLayer := range archiveLayers {
		tarPath := filepath.Join(r.tmpDir, fmt.Sprintf("archive-v1-%03d.tar.gz", i))
		if err := writeTarLayer(tarPath, archiveLayer.files); err != nil {
			return fmt.Errorf("unable to create tar layer %q: %w", archiveLayer.title, err)
		}

		layer, err := r.store.Add(ctx, archiveLayer.title, tarLayerMediaType, tarPath)
		if err != nil {
			return fmt.Errorf("unable to add tar layer %q: %w", archiveLayer.title, err)
		}
		layer.Annotations[orasfile.AnnotationUnpack] = "true"
		layers = append(layers, layer)
	}

	descriptor, err := oras.Pack(ctx, r.store, defaultArtifactType, layers, oras.PackOptions{
		PackImageManifest: true,
		ManifestAnnotations: map[string]string{
			v1.AnnotationCreated: "1970-01-01T00:00:00Z",
		},
	})
	if err != nil {
		return fmt.Errorf("unable to pack manifest: %w", err)
	}

	if err := r.store.Tag(ctx, descriptor, defaultLocalTag); err != nil {
		return fmt.Errorf("unable to tag manifest: %w", err)
	}

	_, err = r.store.Resolve(ctx, defaultLocalTag)
	if err != nil {
		return fmt.Errorf("unable to resolve tag: %w", err)
	}
	log.Info("found tag", zap.String("tag", defaultLocalTag))

	return nil
}

func groupArchiveLayers(files []FileRef) []archiveLayer {
	if len(files) == 0 {
		return nil
	}

	filesByTitle := make(map[string][]FileRef)
	for _, file := range files {
		title := tarLayerTitle
		if topLevel, _, ok := strings.Cut(file.RelPath, "/"); ok {
			title = topLevel
		}
		filesByTitle[title] = append(filesByTitle[title], file)
	}
	if len(filesByTitle) > maxArchiveLayers {
		return []archiveLayer{{title: tarLayerTitle, files: files}}
	}

	titles := make([]string, 0, len(filesByTitle))
	for title := range filesByTitle {
		titles = append(titles, title)
	}
	sort.Strings(titles)

	layers := make([]archiveLayer, 0, len(titles))
	for _, title := range titles {
		layers = append(layers, archiveLayer{title: title, files: filesByTitle[title]})
	}
	return layers
}
