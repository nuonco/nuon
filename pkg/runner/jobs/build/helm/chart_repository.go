package helm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"helm.sh/helm/v4/pkg/chart/loader/archive"
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/registry"
	"helm.sh/helm/v4/pkg/repo/v1"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
	"sigs.k8s.io/yaml"
)

func newChartRepositoryClient() *http.Client {
	return &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: time.Second,
		},
	}
}

func fetchChartData(ctx context.Context, client *http.Client, address string) ([]byte, error) {
	u, err := url.Parse(address)
	if err != nil {
		return nil, err
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("chart downloads require an explicit HTTP(S) URL, not a Helm plugin or local path: %q", address)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/gzip,application/octet-stream,application/yaml")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chart download returned %s", resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, archive.MaxDecompressedChartSize+1))
	if err == nil && int64(len(data)) > archive.MaxDecompressedChartSize {
		err = fmt.Errorf("chart download exceeds Helm chart size limit")
	}
	return data, err
}

func fetchRepositoryChart(ctx context.Context, client *http.Client, repository, name, version string) ([]byte, error) {
	return newChartRepositoryFetcher(client).fetch(ctx, repository, name, version)
}

// A refresh shares downloads, not resolved dependencies: each declaration must
// still resolve its own version constraint and validate the returned chart.
type chartRepositoryFetcher struct {
	client    *http.Client
	indexes   map[string]*repo.IndexFile
	archives  map[string][]byte
	ociLayers map[string][]byte
}

func newChartRepositoryFetcher(client *http.Client) *chartRepositoryFetcher {
	return &chartRepositoryFetcher{
		client: client, indexes: map[string]*repo.IndexFile{},
		archives: map[string][]byte{}, ociLayers: map[string][]byte{},
	}
}

func (f *chartRepositoryFetcher) fetch(ctx context.Context, repository, name, version string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := chartComponent(name); err != nil {
		return nil, err
	}
	if registry.IsOCI(repository) {
		return f.fetchOCI(ctx, strings.TrimSuffix(repository, "/")+"/"+name, version)
	}
	for old, replacement := range map[string]string{
		"//kubernetes-charts.storage.googleapis.com":           "https://charts.helm.sh/stable",
		"//kubernetes-charts-incubator.storage.googleapis.com": "https://charts.helm.sh/incubator",
	} {
		if strings.Contains(repository, old) {
			return nil, fmt.Errorf("repo %q is no longer available; try %q instead", repository, replacement)
		}
	}
	indexURL, err := repo.ResolveReferenceURL(repository, "index.yaml")
	if err != nil {
		return nil, err
	}
	index := f.indexes[indexURL]
	if index == nil {
		data, err := fetchChartData(ctx, f.client, indexURL)
		if err != nil {
			return nil, err
		}
		index, err = chartRepositoryIndex(data)
		if err != nil {
			return nil, err
		}
		f.indexes[indexURL] = index
	}
	entry, err := index.Get(name, version)
	if err != nil {
		return nil, err
	}
	if len(entry.URLs) == 0 {
		return nil, fmt.Errorf("chart %q has no downloadable URLs", name)
	}
	address, err := repo.ResolveReferenceURL(repository, entry.URLs[0])
	if err != nil {
		return nil, err
	}
	if data, ok := f.archives[address]; ok {
		return data, nil
	}
	data, err := fetchChartData(ctx, f.client, address)
	if err == nil {
		f.archives[address] = data
	}
	return data, err
}

// Helm exposes IndexFile.Get/SortEntries but only a path-based index loader.
func chartRepositoryIndex(data []byte) (*repo.IndexFile, error) {
	if len(data) == 0 {
		return nil, repo.ErrEmptyIndexYaml
	}
	index := &repo.IndexFile{}
	var err error
	if json.Valid(data) {
		err = json.Unmarshal(data, index)
	} else {
		err = yaml.UnmarshalStrict(data, index)
	}
	if err != nil {
		return nil, err
	}
	if index.APIVersion == "" {
		return nil, repo.ErrNoAPIVersion
	}
	for name, versions := range index.Entries {
		valid := versions[:0]
		for _, entry := range versions {
			if entry == nil {
				continue
			}
			if entry.Metadata == nil {
				entry.Metadata = &chartv2.Metadata{}
			}
			if entry.APIVersion == "" {
				entry.APIVersion = chartv2.APIVersionV1
			}
			if err := entry.Validate(); err != nil && !strings.HasPrefix(err.Error(), "validation: more than one dependency with name or alias") {
				continue
			}
			valid = append(valid, entry)
		}
		index.Entries[name] = valid
	}
	index.SortEntries()
	return index, nil
}

func (f *chartRepositoryFetcher) fetchOCI(ctx context.Context, reference, version string) ([]byte, error) {
	// registry.NewClient always opens Helm and Docker credentials, even with an
	// injected authorizer. ORAS remote repositories have no filesystem defaults.
	r, err := remote.NewRepository(strings.TrimPrefix(reference, "oci://"))
	if err != nil {
		return nil, err
	}
	r.PlainHTTP = true // Retain this handler's existing plain-HTTP OCI behavior.
	ociClient := *f.client
	ociClient.Transport = retry.NewTransport(f.client.Transport)
	r.Client = &auth.Client{Client: &ociClient, Cache: auth.NewCache()}
	tag := r.Reference.Reference
	if version != "" {
		if tag != "" && tag != version {
			return nil, fmt.Errorf("chart reference and version mismatch")
		}
		tag = version
	}
	if r.Reference.Reference != "" && strings.Contains(r.Reference.Reference, ":") {
		// Digests are already immutable references, not version constraints.
		tag = r.Reference.Reference
	} else if _, err := semver.NewVersion(tag); err != nil {
		var versions semver.Collection
		if err := r.Tags(ctx, "", func(tags []string) error {
			for _, tag := range tags {
				if v, err := semver.StrictNewVersion(strings.ReplaceAll(tag, "_", "+")); err == nil {
					versions = append(versions, v)
				}
			}
			return nil
		}); err != nil {
			return nil, err
		}
		sort.Sort(sort.Reverse(versions))
		tags := make([]string, len(versions))
		for i, v := range versions {
			tags[i] = v.String()
		}
		tag, err = registry.GetTagMatchingVersionOrConstraint(tags, tag)
		if err != nil {
			return nil, err
		}
	}
	desc, err := r.Resolve(ctx, strings.ReplaceAll(tag, "+", "_"))
	if err != nil {
		return nil, err
	}
	if desc.Size < 0 || desc.Size > archive.MaxDecompressedFileSize {
		return nil, fmt.Errorf("OCI chart manifest exceeds Helm file size limit")
	}
	data, err := content.FetchAll(ctx, r, desc)
	if err != nil {
		return nil, err
	}
	var manifest ocispec.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	if manifest.Config.MediaType != registry.ConfigMediaType {
		return nil, fmt.Errorf("OCI manifest does not contain a Helm config")
	}
	if manifest.Config.Size < 0 || manifest.Config.Size > archive.MaxDecompressedFileSize {
		return nil, fmt.Errorf("OCI chart config exceeds Helm file size limit")
	}
	config, err := content.FetchAll(ctx, r, manifest.Config)
	if err != nil {
		return nil, err
	}
	var metadata chartv2.Metadata
	if err := json.Unmarshal(config, &metadata); err != nil {
		return nil, err
	}
	for _, layer := range manifest.Layers {
		if layer.MediaType == registry.ChartLayerMediaType || layer.MediaType == registry.LegacyChartLayerMediaType {
			if layer.Size < 0 || layer.Size > archive.MaxDecompressedChartSize {
				return nil, fmt.Errorf("OCI chart layer exceeds Helm chart size limit")
			}
			address := "oci://" + r.Reference.Registry + "/" + r.Reference.Repository + "@" + layer.Digest.String()
			if data, ok := f.ociLayers[address]; ok {
				if int64(len(data)) != layer.Size {
					return nil, fmt.Errorf("OCI chart layer size does not match its descriptor")
				}
				return data, nil
			}
			data, err := content.FetchAll(ctx, r, layer)
			if err == nil {
				f.ociLayers[address] = data
			}
			return data, err
		}
	}
	return nil, fmt.Errorf("OCI manifest does not contain a Helm chart layer")
}
