/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package helm

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"time"

	"helm.sh/helm/v4/pkg/chart/common"
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
	"sigs.k8s.io/yaml"
)

// Adapted from Helm v4.1.4 chart/v2/util/save.go: its serializer is unexported
// and Save opens host paths. Keep the wire format, but accept only a writer.
func writeChartArchive(out io.Writer, chart *chartv2.Chart) error {
	if err := chart.Validate(); err != nil {
		return fmt.Errorf("chart validation: %w", err)
	}
	if err := checkChartPaths(chart); err != nil {
		return err
	}
	z := gzip.NewWriter(out)
	z.Extra = []byte("+aHR0cHM6Ly95b3V0dS5iZS96OVV6MWljandyTQo=")
	z.Comment = "Helm"
	tw := tar.NewWriter(z)
	err := writeChartContents(tw, chart, "")
	return errors.Join(err, tw.Close(), z.Close())
}

func writeChartContents(out *tar.Writer, chart *chartv2.Chart, prefix string) error {
	base := path.Join(prefix, chart.Name())
	metadata := *chart.Metadata
	if metadata.APIVersion == chartv2.APIVersionV1 {
		metadata.Dependencies = nil
	}
	data, err := yaml.Marshal(&metadata)
	if err != nil {
		return err
	}
	if err := writeChartFile(out, path.Join(base, "Chart.yaml"), data, chart.ModTime); err != nil {
		return err
	}
	if metadata.APIVersion == chartv2.APIVersionV2 && chart.Lock != nil {
		data, err := yaml.Marshal(chart.Lock)
		if err != nil {
			return err
		}
		if err := writeChartFile(out, path.Join(base, "Chart.lock"), data, chart.Lock.Generated); err != nil {
			return err
		}
	}
	for _, file := range chart.Raw {
		if file.Name == "values.yaml" {
			if err := writeChartFile(out, path.Join(base, file.Name), file.Data, file.ModTime); err != nil {
				return err
			}
		}
	}
	if chart.Schema != nil {
		if !json.Valid(chart.Schema) {
			return fmt.Errorf("invalid JSON in values.schema.json")
		}
		if err := writeChartFile(out, path.Join(base, "values.schema.json"), chart.Schema, chart.SchemaModTime); err != nil {
			return err
		}
	}
	for _, files := range [][]*common.File{chart.Templates, chart.Files} {
		for _, file := range files {
			if err := chartFilePath(file.Name); err != nil {
				return err
			}
			if err := writeChartFile(out, path.Join(base, file.Name), file.Data, file.ModTime); err != nil {
				return err
			}
		}
	}
	for _, dep := range chart.Dependencies() {
		if err := writeChartContents(out, dep, path.Join(base, "charts")); err != nil {
			return err
		}
	}
	return nil
}

func writeChartFile(out *tar.Writer, name string, data []byte, modTime time.Time) error {
	if modTime.IsZero() {
		modTime = time.Now()
	}
	if err := out.WriteHeader(&tar.Header{Name: name, Mode: 0644, Size: int64(len(data)), ModTime: modTime}); err != nil {
		return err
	}
	_, err := out.Write(data)
	return err
}
