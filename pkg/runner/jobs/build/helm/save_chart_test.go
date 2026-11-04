package helm

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v4/pkg/chart/common"
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
	chartutil "helm.sh/helm/v4/pkg/chart/v2/util"
)

func TestRootedChartSerializationMatchesHelm414(t *testing.T) {
	for _, apiVersion := range []string{"v1", "v2"} {
		t.Run(apiVersion, func(t *testing.T) {
			modTime := time.Unix(1700000000, 0)
			chart := &chartv2.Chart{
				Metadata: &chartv2.Metadata{
					APIVersion: apiVersion, Name: "acme", Version: "1.2.3+meta", AppVersion: "4.5.6", Type: "application",
					Annotations: map[string]string{"example.com/source": "retained"}, Sources: []string{"https://example.com/acme"},
					Dependencies: []*chartv2.Dependency{{Name: "child", Version: "0.1.0", Repository: "file://../child"}},
				},
				ModTime:   modTime,
				Raw:       []*common.File{{Name: "values.yaml", Data: []byte("\xef\xbb\xbfvalue: retained\n"), ModTime: modTime}},
				Templates: []*common.File{{Name: "templates/config.yaml", Data: []byte("\xef\xbb\xbfapiVersion: v1\nkind: ConfigMap\n"), ModTime: modTime}},
				Files:     []*common.File{{Name: "README.md", Data: []byte("readme"), ModTime: modTime}},
				Schema:    []byte(`{"type":"object"}`), SchemaModTime: modTime,
				Lock: &chartv2.Lock{Generated: modTime, Digest: "sha256:fixture", Dependencies: []*chartv2.Dependency{{Name: "child", Version: "0.1.0", Repository: "file://../child"}}},
			}
			child := &chartv2.Chart{
				Metadata: &chartv2.Metadata{APIVersion: "v2", Name: "child", Version: "0.1.0"}, ModTime: modTime,
				Files: []*common.File{{Name: "crds/example.yaml", Data: []byte("apiVersion: v1\nkind: ConfigMap\n"), ModTime: modTime}},
			}
			chart.AddDependency(child)
			if apiVersion == "v1" {
				chart.Files = append(chart.Files,
					&common.File{Name: "requirements.yaml", Data: []byte("dependencies:\n- name: child\n  version: 0.1.0\n  repository: file://../child\n"), ModTime: modTime},
					&common.File{Name: "requirements.lock", Data: []byte("dependencies:\n- name: child\n  version: 0.1.0\n  repository: file://../child\n"), ModTime: modTime},
				)
			}
			var got bytes.Buffer
			require.NoError(t, writeChartArchive(&got, chart))
			path, err := chartutil.Save(chart, t.TempDir())
			require.NoError(t, err)
			want, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, want, got.Bytes(), "metadata, schema, raw values/BOM, files, subcharts, tar headers and gzip framing must match Helm")
			require.Len(t, chart.Metadata.Dependencies, 1, "serialization must not mutate v1 metadata")
			fs := afero.NewMemMapFs()
			published, err := publishChart(fs, chart)
			require.NoError(t, err)
			require.Equal(t, "acme", published.Name())
			data, err := afero.ReadFile(fs, defaultChartPackageFilename)
			require.NoError(t, err)
			require.Equal(t, want, data)
			entries, err := afero.ReadDir(fs, ".")
			require.NoError(t, err)
			require.Len(t, entries, 1, "serializer scratch files must be removed")
		})
	}
}

type failingChartWriter struct{}

func (failingChartWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestChartSerializationErrorsAndZeroTimestamps(t *testing.T) {
	chart := &chartv2.Chart{Metadata: &chartv2.Metadata{APIVersion: "v2", Name: "acme", Version: "1.0.0"}}
	require.ErrorIs(t, writeChartArchive(failingChartWriter{}, chart), io.ErrClosedPipe)
	var data bytes.Buffer
	start := time.Now().Round(time.Second)
	require.NoError(t, writeChartArchive(&data, chart))
	z, err := gzip.NewReader(bytes.NewReader(data.Bytes()))
	require.NoError(t, err)
	defer z.Close()
	header, err := tar.NewReader(z).Next()
	require.NoError(t, err)
	require.False(t, header.ModTime.Before(start))
	require.False(t, header.ModTime.After(time.Now().Round(time.Second)))
	chart.Schema = []byte("invalid JSON")
	fs := afero.NewMemMapFs()
	_, err = publishChart(fs, chart)
	require.ErrorContains(t, err, "invalid JSON")
	entries, err := afero.ReadDir(fs, ".")
	require.NoError(t, err)
	require.Empty(t, entries)
	_, err = fs.Stat(defaultChartPackageFilename)
	require.True(t, errors.Is(err, os.ErrNotExist))
}
