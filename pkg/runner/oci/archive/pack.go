package ociarchive

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	orasfile "oras.land/oras-go/v2/content/file"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/op"
	"github.com/nuonco/nuon/pkg/sourcefs"
)

const (
	defaultArtifactType string = "artifact/nuon"
	defaultLocalTag     string = "latest"
	tarLayerMediaType   string = "application/vnd.nuon.archive.layer.v1.tar+gzip"
	tarLayerTitle       string = "."
	maxArchiveLayers           = 400
)

// FileRef maps an input in the supplied source filesystem to an archive name.
// AbsPath may be source-relative or absolute within the source root; RelPath
// must be a canonical, slash-separated relative archive path.
type FileRef struct {
	AbsPath  string `mapstructure:"abs_path,omitempty"`
	RelPath  string `mapstructure:"rel_path,omitempty"`
	FileType string `mapstructure:"file_type,omitempty"`
}

type archiveLayer struct {
	title string
	files []FileRef
}

func (r *archive) Pack(ctx context.Context, log *zap.Logger, source *sourcefs.FS, filePaths []FileRef) (retErr error) {
	opCtx, end := op.Tool(ctx, "oci", "pack")
	ctx = opCtx
	defer func() { end(retErr) }()
	if l, err := pkgctx.Logger(ctx); err == nil && l != nil {
		log = l
	}
	if source == nil {
		return fmt.Errorf("archive source filesystem is required")
	}
	if r.store == nil {
		return fmt.Errorf("archive is not initialized")
	}
	if err := r.store.checkOpen(); err != nil {
		return err
	}

	archiveLayers := groupArchiveLayers(filePaths)
	layers := make([]v1.Descriptor, 0, len(archiveLayers))
	for i, archiveLayer := range archiveLayers {
		tarPath := fmt.Sprintf("layers/archive-v1-%03d.tar.gz", i)
		out, err := r.workFS.OpenFile(tarPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return fmt.Errorf("unable to create tar layer %q: %w", archiveLayer.title, err)
		}
		writeErr := writeTarLayer(out, source, archiveLayer.files)
		var layer v1.Descriptor
		if writeErr == nil {
			layer, writeErr = r.addLayer(ctx, archiveLayer.title, out)
		}
		closeErr := out.Close()
		if writeErr != nil {
			return fmt.Errorf("unable to create tar layer %q: %w", archiveLayer.title, writeErr)
		}
		if closeErr != nil {
			return fmt.Errorf("unable to close tar layer %q: %w", archiveLayer.title, closeErr)
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

	// Packing publishes the CAS without extracting a second copy of the inputs.
	if err := r.store.contentStore.Tag(ctx, descriptor, defaultLocalTag); err != nil {
		return fmt.Errorf("unable to tag manifest: %w", err)
	}

	_, err = r.store.Resolve(ctx, defaultLocalTag)
	if err != nil {
		return fmt.Errorf("unable to resolve tag: %w", err)
	}
	log.Info("found tag", zap.String("tag", defaultLocalTag))

	return nil
}

func (r *archive) addLayer(ctx context.Context, title string, input io.ReadSeeker) (v1.Descriptor, error) {
	if _, err := input.Seek(0, io.SeekStart); err != nil {
		return v1.Descriptor{}, err
	}
	digester := digest.Canonical.Digester()
	size, err := io.Copy(digester.Hash(), &contextReader{ctx: ctx, reader: input})
	if err != nil {
		return v1.Descriptor{}, err
	}
	if _, err := input.Seek(0, io.SeekStart); err != nil {
		return v1.Descriptor{}, err
	}
	desc := v1.Descriptor{
		MediaType:   tarLayerMediaType,
		Digest:      digester.Digest(),
		Size:        size,
		Annotations: map[string]string{v1.AnnotationTitle: title},
	}
	if err := r.store.Push(ctx, desc, &contextReader{ctx: ctx, reader: input}); err != nil {
		return v1.Descriptor{}, err
	}
	return desc, nil
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
