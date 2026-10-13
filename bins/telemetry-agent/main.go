package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/telemetryexport"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	if err := run(logger); err != nil {
		logger.Error("telemetry agent stopped", zap.Error(err))
		os.Exit(1)
	}
}

func run(logger *zap.Logger) error {
	apiURL := flag.String("api-url", "", "Nuon runner API base URL")
	installID := flag.String("install-id", "", "Nuon install ID")
	tokenFile := flag.String("token-file", "", "Mounted opaque Nuon API-token file")
	dataDir := flag.String("data-dir", "/var/lib/nuon/telemetry-agent", "Queue and relay credential directory")
	binary := flag.String("collector-binary", "/bin/nuon-otelcol", "OpenTelemetry Collector executable")
	healthAddress := flag.String("health-address", "0.0.0.0:13133", "Agent health listen address")
	grpcAddress := flag.String("otlp-grpc-address", "0.0.0.0:4317", "OTLP gRPC listen address")
	httpAddress := flag.String("otlp-http-address", "0.0.0.0:4318", "OTLP HTTP listen address")
	collectorHealthAddress := flag.String("collector-health-address", "127.0.0.1:13134", "Private child health listen address")
	allowHTTP := flag.Bool("allow-insecure-api", false, "Allow a trusted local development runner API over HTTP")
	flag.Parse()
	parsed, err := url.Parse(*apiURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && !(parsed.Scheme == "http" && *allowHTTP)) || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("api-url must be an HTTPS base URL without userinfo, query, or fragment (HTTP requires allow-insecure-api)")
	}
	if *installID == "" || strings.ContainsAny(*installID, "/ \t\r\n") || *tokenFile == "" {
		return fmt.Errorf("install-id and token-file are required")
	}
	if _, err := os.Stat(*binary); err != nil {
		return fmt.Errorf("collector executable unavailable: %w", err)
	}
	client, err := nuonrunner.New(nuonrunner.WithURL(*apiURL), nuonrunner.WithAuthTokenFile(*tokenFile), nuonrunner.WithRequestTimeout(15*time.Second))
	if err != nil {
		return err
	}
	config := telemetryexport.DefaultConfig(*dataDir)
	config.GRPCAddress, config.HTTPAddress, config.HealthAddress = *grpcAddress, *httpAddress, *collectorHealthAddress
	supervisor := telemetryexport.NewVendor(telemetryexport.Options{
		InstallID: *installID, Binary: *binary, Config: config, Client: client, Logger: logger,
		FetchSettings: func(ctx context.Context) (telemetryexport.Settings, error) {
			response, err := client.GetInstallTelemetryCollectorSettings(ctx, *installID)
			if err != nil {
				return telemetryexport.Settings{}, err
			}
			if response == nil {
				return telemetryexport.Settings{}, fmt.Errorf("telemetry settings response is empty")
			}
			return telemetryexport.Settings{
				Enabled:            response.Enabled,
				RelayEndpoint:      response.RelayEndpoint,
				ResourceAttributes: response.ResourceAttributes,
			}, nil
		},
	})
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if !supervisor.Ready() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: time.Minute}
	listener, err := net.Listen("tcp", *healthAddress)
	if err != nil {
		return err
	}
	serverErr := make(chan error, 1)
	go func() { serverErr <- server.Serve(listener) }()
	done := make(chan struct{})
	go func() { supervisor.Run(ctx); close(done) }()
	select {
	case <-ctx.Done():
	case err = <-serverErr:
		cancel()
	}
	// Run stops the child and removes relay credentials before returning.
	<-done
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = server.Shutdown(shutdownCtx)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
