package helm

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"

	"helm.sh/helm/v3/pkg/chart"
	rspb "helm.sh/helm/v3/pkg/release"
)

var b64 = base64.StdEncoding

var magicGzip = []byte{0x1f, 0x8b, 0x08}

var systemLabels = []string{"name", "owner", "status", "version", "createdAt", "modifiedAt"}

// Release describes a deployment of a chart, together with the chart
// and the variables used to deploy that chart.
// we had to make our own copy of the struct to expose the Labels field
type Release struct {
	// Name is the name of the release
	Name string `json:"name,omitempty"`
	// Info provides information about a release
	Info *rspb.Info `json:"info,omitempty"`
	// Chart is the chart that was released.
	Chart *chart.Chart `json:"chart,omitempty"`
	// Config is the set of extra Values added to the chart.
	// These values override the default values inside of the chart.
	Config map[string]interface{} `json:"config,omitempty"`
	// Manifest is the string representation of the rendered template.
	Manifest string `json:"manifest,omitempty"`
	// Hooks are all of the hooks declared for this release.
	Hooks []*rspb.Hook `json:"hooks,omitempty"`
	// Version is an int which represents the revision of the release.
	Version int `json:"version,omitempty"`
	// Namespace is the kubernetes namespace of the release.
	Namespace string `json:"namespace,omitempty"`
	// Labels of the release.
	Labels map[string]string `json:"labels,omitempty"`
}

// EncodeRelease encodes a release returning a base64 encoded
// gzipped string representation, or error.
func EncodeRelease(rls *Release) (string, error) {
	b, err := json.Marshal(rls)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return "", err
	}
	if _, err = w.Write(b); err != nil {
		return "", err
	}
	w.Close()

	return b64.EncodeToString(buf.Bytes()), nil
}

// DecodeRelease decodes the bytes of data into a release
// type. Data must contain a base64 encoded gzipped string of a
// valid release, otherwise an error is returned.
func DecodeRelease(data string) (*Release, error) {
	// base64 decode string
	b, err := b64.DecodeString(data)
	if err != nil {
		return nil, err
	}

	// For backwards compatibility with releases that were stored before
	// compression was introduced we skip decompression if the
	// gzip magic header is not found
	if len(b) > 3 && bytes.Equal(b[0:3], magicGzip) {
		r, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		defer r.Close()
		b2, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		b = b2
	}

	var rls Release
	// unmarshal release object bytes
	if err := json.Unmarshal(b, &rls); err != nil {
		return nil, err
	}
	return &rls, nil
}

// Checks if label is system
func isSystemLabel(key string) bool {
	for _, v := range GetSystemLabels() {
		if key == v {
			return true
		}
	}
	return false
}

// Removes system labels from labels map
func FilterSystemLabels(lbs map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range lbs {
		if !isSystemLabel(k) {
			result[k] = v
		}
	}
	return result
}

// Checks if labels array contains system labels
func ContainsSystemLabels(lbs map[string]string) bool {
	for k := range lbs {
		if isSystemLabel(k) {
			return true
		}
	}
	return false
}

func GetSystemLabels() []string {
	return systemLabels
}
