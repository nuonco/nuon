package supportsnapshot

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"

	"github.com/klauspost/compress/zstd"
)

const (
	MaxArchiveSize  = int64(512 << 20)
	MaxExpandedSize = int64(128 << 20)
	MaxEntries      = 16
)

type Archive struct {
	Manifest     Manifest
	Registration []byte
	Snapshot     Snapshot
	Collection   CollectionReport
}

func Write(w io.Writer, snapshot Snapshot, producer Producer) (Manifest, error) {
	if snapshot.SchemaVersion != SchemaVersion {
		return Manifest{}, fmt.Errorf("unsupported support snapshot schema version %d", snapshot.SchemaVersion)
	}
	if err := snapshot.Registration.Validate(); err != nil {
		return Manifest{}, fmt.Errorf("invalid installation registration: %w", err)
	}
	registration, err := json.Marshal(snapshot.Registration)
	if err != nil {
		return Manifest{}, fmt.Errorf("encode installation registration: %w", err)
	}
	snapshotRaw, err := json.Marshal(snapshot)
	if err != nil {
		return Manifest{}, fmt.Errorf("encode support snapshot: %w", err)
	}
	collection, err := json.Marshal(snapshot.Collection)
	if err != nil {
		return Manifest{}, fmt.Errorf("encode collection report: %w", err)
	}
	files := map[string]fileEntry{
		RegistrationPath: {mediaType: "application/json", data: registration},
		SnapshotPath:     {mediaType: "application/vnd.nuon.customer-managed-support-snapshot+json", data: snapshotRaw},
		CollectionPath:   {mediaType: "application/vnd.nuon.customer-managed-support-collection+json", data: collection},
	}
	manifest := Manifest{
		SchemaVersion: SchemaVersion,
		CapturedAt:    snapshot.CapturedAt.UTC(),
		Producer:      producer,
		Registration:  snapshot.Registration.RegistrationID,
		BundleDigest:  snapshot.Registration.BundleDigest,
	}
	for name, file := range files {
		sum := sha256.Sum256(file.data)
		manifest.Entries = append(manifest.Entries, ManifestEntry{
			Path: name, MediaType: file.mediaType, SchemaVersion: SchemaVersion,
			Size: int64(len(file.data)), SHA256: hex.EncodeToString(sum[:]),
		})
	}
	sort.Slice(manifest.Entries, func(i, j int) bool { return manifest.Entries[i].Path < manifest.Entries[j].Path })
	manifestRaw, err := json.Marshal(manifest)
	if err != nil {
		return Manifest{}, fmt.Errorf("encode support snapshot manifest: %w", err)
	}
	encoder, err := zstd.NewWriter(w, zstd.WithEncoderLevel(zstd.SpeedDefault))
	if err != nil {
		return Manifest{}, fmt.Errorf("create zstd encoder: %w", err)
	}
	tw := tar.NewWriter(encoder)
	if err := writeTarFile(tw, ManifestPath, manifestRaw); err != nil {
		return Manifest{}, closeWriters(tw, encoder, err)
	}
	for _, entry := range manifest.Entries {
		if err := writeTarFile(tw, entry.Path, files[entry.Path].data); err != nil {
			return Manifest{}, closeWriters(tw, encoder, err)
		}
	}
	if err := tw.Close(); err != nil {
		encoder.Close()
		return Manifest{}, fmt.Errorf("close support snapshot tar: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return Manifest{}, fmt.Errorf("close support snapshot encoder: %w", err)
	}
	return manifest, nil
}

func Read(r io.Reader) (*Archive, error) {
	decoder, err := zstd.NewReader(io.LimitReader(r, MaxArchiveSize+1), zstd.WithDecoderMaxMemory(uint64(MaxExpandedSize)))
	if err != nil {
		return nil, fmt.Errorf("open support snapshot: %w", err)
	}
	defer decoder.Close()
	tr := tar.NewReader(decoder)
	files := make(map[string][]byte)
	var expanded int64
	for count := 0; ; count++ {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read support snapshot archive: %w", err)
		}
		if count >= MaxEntries {
			return nil, fmt.Errorf("support snapshot contains more than %d entries", MaxEntries)
		}
		if header.Typeflag != tar.TypeReg || !safeArchivePath(header.Name) {
			return nil, fmt.Errorf("support snapshot entry %q is not a safe regular file", header.Name)
		}
		if _, exists := files[header.Name]; exists {
			return nil, fmt.Errorf("support snapshot contains duplicate entry %q", header.Name)
		}
		if header.Size < 0 || expanded+header.Size > MaxExpandedSize {
			return nil, fmt.Errorf("support snapshot exceeds the %d byte expanded limit", MaxExpandedSize)
		}
		data, err := io.ReadAll(io.LimitReader(tr, header.Size+1))
		if err != nil {
			return nil, fmt.Errorf("read support snapshot entry %q: %w", header.Name, err)
		}
		if int64(len(data)) != header.Size {
			return nil, fmt.Errorf("support snapshot entry %q size mismatch", header.Name)
		}
		expanded += header.Size
		files[header.Name] = data
	}
	var manifest Manifest
	if err := json.Unmarshal(files[ManifestPath], &manifest); err != nil {
		return nil, fmt.Errorf("decode support snapshot manifest: %w", err)
	}
	if manifest.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("unsupported support snapshot schema version %d", manifest.SchemaVersion)
	}
	if err := validateEntries(manifest, files); err != nil {
		return nil, err
	}
	archive := &Archive{Manifest: manifest, Registration: files[RegistrationPath]}
	if err := json.Unmarshal(files[SnapshotPath], &archive.Snapshot); err != nil {
		return nil, fmt.Errorf("decode support snapshot: %w", err)
	}
	if err := json.Unmarshal(files[CollectionPath], &archive.Collection); err != nil {
		return nil, fmt.Errorf("decode support snapshot collection report: %w", err)
	}
	if archive.Snapshot.SchemaVersion != SchemaVersion || archive.Collection.SchemaVersion != SchemaVersion {
		return nil, errors.New("support snapshot sections do not match the manifest schema version")
	}
	if err := archive.Snapshot.Registration.Validate(); err != nil {
		return nil, fmt.Errorf("validate support snapshot registration: %w", err)
	}
	if archive.Snapshot.Registration.RegistrationID != manifest.Registration || archive.Snapshot.Registration.BundleDigest != manifest.BundleDigest {
		return nil, errors.New("support snapshot identity does not match its manifest")
	}
	return archive, nil
}

type fileEntry struct {
	mediaType string
	data      []byte
}

func writeTarFile(tw *tar.Writer, name string, data []byte) error {
	header := &tar.Header{Name: name, Mode: 0o600, Size: int64(len(data)), Typeflag: tar.TypeReg}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err := io.Copy(tw, bytes.NewReader(data))
	return err
}

func closeWriters(tw *tar.Writer, encoder *zstd.Encoder, cause error) error {
	_ = tw.Close()
	_ = encoder.Close()
	return cause
}

func safeArchivePath(name string) bool {
	return name != "" && path.Clean(name) == name && !strings.HasPrefix(name, "/") && !strings.HasPrefix(name, "../")
}

func validateEntries(manifest Manifest, files map[string][]byte) error {
	wanted := map[string]bool{RegistrationPath: true, SnapshotPath: true, CollectionPath: true}
	if len(manifest.Entries) != len(wanted) {
		return errors.New("support snapshot manifest has an unexpected entry count")
	}
	seen := make(map[string]bool, len(manifest.Entries))
	for _, entry := range manifest.Entries {
		if !wanted[entry.Path] || seen[entry.Path] {
			return fmt.Errorf("support snapshot manifest contains unexpected entry %q", entry.Path)
		}
		data, ok := files[entry.Path]
		if !ok || int64(len(data)) != entry.Size {
			return fmt.Errorf("support snapshot entry %q is missing or has the wrong size", entry.Path)
		}
		sum := sha256.Sum256(data)
		if entry.SHA256 != hex.EncodeToString(sum[:]) {
			return fmt.Errorf("support snapshot entry %q failed SHA-256 verification", entry.Path)
		}
		seen[entry.Path] = true
	}
	if len(files) != len(wanted)+1 {
		return errors.New("support snapshot archive contains entries absent from its manifest")
	}
	return nil
}
