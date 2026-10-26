package version

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func configuredConfig(apiURL string) *config.Config {
	v := viper.New()
	v.Set("api_url", apiURL)
	return &config.Config{Viper: v, APIURL: apiURL}
}

func TestVersionReportUsesSDK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/version":
			_, _ = w.Write([]byte(`{"version":"0.19.1109"}`))
		case "/v1/general/cli-config":
			_, _ = w.Write([]byte(`{"min_cli_version":"0.19.1102"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	api, err := nuon.New(nuon.WithURL(srv.URL))
	require.NoError(t, err)

	got := New(api, configuredConfig(srv.URL)).collect(context.Background())
	assert.Equal(t, Version, got.client)
	assert.Equal(t, "0.19.1109", got.ctlAPI)
	assert.Equal(t, "0.19.1102", got.minCLI)
	assert.Equal(t, "client: "+Version+"\nctl-api: 0.19.1109\nmin_cli_supported_version: 0.19.1102\n", got.text())

	out := got.json()
	assert.Equal(t, Version, out["version"])
	assert.Equal(t, "0.19.1109", out["ctl_api_version"])
	assert.Equal(t, "0.19.1102", out["min_cli_supported_version"])
}

func TestVersionReportOmitsUnsetMinCLI(t *testing.T) {
	api := &stubAPI{version: &nuon.APIVersion{Version: "0.19.1109"}}
	got := New(api, configuredConfig("https://api.example.com")).collect(context.Background())
	assert.Equal(t, "client: "+Version+"\nctl-api: 0.19.1109\n", got.text())
	_, ok := got.json()["min_cli_supported_version"]
	assert.False(t, ok)
}

func TestVersionReportOmitsFailedLookup(t *testing.T) {
	api := &stubAPI{versionErr: errors.New("dial"), cfgErr: errors.New("dial")}
	got := New(api, configuredConfig("https://api.example.com")).collect(context.Background())
	assert.Equal(t, "client: "+Version+"\n", got.text())
	_, hasCtl := got.json()["ctl_api_version"]
	_, hasMin := got.json()["min_cli_supported_version"]
	assert.False(t, hasCtl)
	assert.False(t, hasMin)
}

func TestVersionReportSkipsUnconfiguredAPI(t *testing.T) {
	api := &stubAPI{
		version: &nuon.APIVersion{Version: "0.19.1109"},
		cfg:     &models.ServiceCLIConfig{MinCliVersion: "0.19.1102"},
	}
	got := New(api, &config.Config{Viper: viper.New()}).collect(context.Background())
	assert.Equal(t, "client: "+Version+"\n", got.text())
	assert.Zero(t, api.versionCalls)
	assert.Zero(t, api.configCalls)
}

func TestVersionReportLookupTimesOut(t *testing.T) {
	svc := New(&stubAPI{block: true}, configuredConfig("https://api.example.com"))
	done := make(chan info, 1)
	go func() {
		done <- svc.collect(context.Background())
	}()

	select {
	case got := <-done:
		assert.Equal(t, "client: "+Version+"\n", got.text())
	case <-time.After(2 * time.Second):
		t.Fatal("version lookup did not return within 2s")
	}
}

type stubAPI struct {
	nuon.Client
	version      *nuon.APIVersion
	versionErr   error
	cfg          *models.ServiceCLIConfig
	cfgErr       error
	block        bool
	versionCalls int
	configCalls  int
}

func (s *stubAPI) GetVersion(ctx context.Context) (*nuon.APIVersion, error) {
	s.versionCalls++
	if s.block {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return s.version, s.versionErr
}

func (s *stubAPI) GetCLIConfig(ctx context.Context) (*models.ServiceCLIConfig, error) {
	s.configCalls++
	if s.block {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return s.cfg, s.cfgErr
}
