package plans

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"

	"github.com/pkg/errors"
)

// CompressPlan compresses plan bytes and returns a base64-encoded string
func CompressPlan(data []byte) (string, error) {
	var zippedBytes bytes.Buffer
	gzipWriter := gzip.NewWriter(&zippedBytes)
	if _, err := gzipWriter.Write(data); err != nil {
		return "", errors.Wrap(err, "failed to write to gzip writer")
	}
	if err := gzipWriter.Close(); err != nil {
		return "", errors.Wrap(err, "failed to close gzip writer")
	}

	encodedString := base64.URLEncoding.EncodeToString(zippedBytes.Bytes())
	return encodedString, nil
}

// DecompressPlan decodes and decompresses a base64-encoded plan content
func DecompressPlan(encodedPlan string) ([]byte, error) {
	// Try to decode using URL-safe encoding
	decodedBytes, err := base64.URLEncoding.DecodeString(encodedPlan)
	if err == nil {
		// Successfully decoded with base64.URLEncoding
	} else {
		// If URL-safe encoding fails, try standard encoding
		decodedBytes, err = base64.StdEncoding.DecodeString(encodedPlan)
		if err != nil {
			// If standard encoding fails, try with RawURLEncoding (no padding)
			decodedBytes, err = base64.RawURLEncoding.DecodeString(encodedPlan)
			if err != nil {
				return nil, errors.Wrap(err, "unable to decode contents from base64")
			}
		}
	}

	// Decompress
	contentsBuffer := bytes.NewReader(decodedBytes)
	reader, err := gzip.NewReader(contentsBuffer)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read contents into gzip reader")
	}
	defer reader.Close()

	decompressedBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decompress contents")
	}

	return decompressedBytes, nil
}
