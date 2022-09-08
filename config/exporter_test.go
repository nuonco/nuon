package config

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:bodyclose,gosec,noctx // NOTE(jdt): these issues were inherited. ignore for now.
func TestMetricsExporter(t *testing.T) {
	var cfg Base
	cfg.ServiceName = "test"
	cfg.ServiceOwner = "owner"
	cfg.SystemPort = 9102

	t.Run("default server", func(t *testing.T) {
		exporter, err := RegisterExporter(cfg)
		require.NoError(t, err)
		require.NotNil(t, exporter)

		t.Run("start", func(t *testing.T) {
			if err = exporter.Start(); err != http.ErrServerClosed {
				require.NoError(t, err)
			}
			<-time.After(50 * time.Millisecond)
		})

		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/%s", cfg.SystemPort, ConfigPath))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		defer resp.Body.Close()

		t.Run("stop", func(t *testing.T) {
			exporter.Stop()

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/%s", cfg.SystemPort, ConfigPath))
			require.Error(t, err)
			assert.Nil(t, resp)
		})
	})

	t.Run("with server", func(t *testing.T) {
		t.Run("no routes", func(t *testing.T) {
			server := &http.Server{
				Addr: fmt.Sprintf(":%d", cfg.SystemPort),
			}
			go func() {
				if err := server.ListenAndServe(); err != http.ErrServerClosed {
					require.NoError(t, server.ListenAndServe())
				}
			}()
			<-time.After(50 * time.Millisecond)

			exporter, err := RegisterExporter(cfg, WithHTTPServer(server))
			require.NoError(t, err)
			require.NotNil(t, exporter)

			t.Run("start", func(t *testing.T) {
				require.NoError(t, exporter.Start())
			})

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/%s", cfg.SystemPort, ConfigPath))
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			t.Run("stop", func(t *testing.T) {
				exporter.Stop()

				resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/%s", cfg.SystemPort, ConfigPath))
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			})
			require.NoError(t, server.Shutdown(context.Background()))
		})

		t.Run("no routes", func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/hi", func(http.ResponseWriter, *http.Request) {})
			server := &http.Server{
				Addr:    fmt.Sprintf(":%d", cfg.SystemPort),
				Handler: mux,
			}
			go func() {
				if err := server.ListenAndServe(); err != http.ErrServerClosed {
					require.NoError(t, server.ListenAndServe())
				}
			}()
			<-time.After(50 * time.Millisecond)

			exporter, err := RegisterExporter(cfg, WithHTTPServer(server))
			require.NoError(t, err)
			require.NotNil(t, exporter)

			t.Run("start", func(t *testing.T) {
				require.NoError(t, exporter.Start())
			})

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/%s", cfg.SystemPort, ConfigPath))
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			t.Run("stop", func(t *testing.T) {
				exporter.Stop()

				resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/%s", cfg.SystemPort, ConfigPath))
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			})
			require.NoError(t, server.Shutdown(context.Background()))
		})
	})
}
