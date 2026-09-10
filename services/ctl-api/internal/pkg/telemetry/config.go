package telemetry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"golang.org/x/net/http/httpguts"
)

type Config struct {
	Endpoint string
	Resource *resource.Resource
}

func NewConfig(cfg *internal.Config) (*Config, error) {
	endpoint := cfg.OTELExporterOTLPEndpoint
	if endpoint != "" {
		u, err := url.Parse(endpoint)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
			return nil, fmt.Errorf("OTEL requires an HTTP(S) OTLP endpoint without userinfo, query, or fragment")
		}
		if protocol := cfg.OTELExporterOTLPProtocol; protocol != "" && protocol != "http/protobuf" {
			return nil, fmt.Errorf("OTEL supports only the http/protobuf protocol")
		}
		if err := validateTransportEnv(u.Scheme); err != nil {
			return nil, err
		}
	}

	rsrc, err := resource.New(context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", cfg.ServiceName),
			attribute.String("service.version", cfg.Version),
			attribute.String("service.instance.id", uuid.NewString()),
			attribute.String("nuon.service.type", cfg.ServiceType),
			attribute.String("nuon.service.deployment", cfg.ServiceDeployment),
		),
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid OTEL_RESOURCE_ATTRIBUTES")
	}
	return &Config{Endpoint: endpoint, Resource: rsrc}, nil
}

func validateTransportEnv(scheme string) error {
	for _, suffix := range []string{"ENDPOINT", "PROTOCOL", "HEADERS", "CERTIFICATE", "CLIENT_CERTIFICATE", "CLIENT_KEY", "TIMEOUT", "COMPRESSION", "INSECURE"} {
		key := "OTEL_EXPORTER_OTLP_METRICS_" + suffix
		if os.Getenv(key) != "" {
			return fmt.Errorf("%s is unsupported; use OTEL_EXPORTER_OTLP_%s", key, suffix)
		}
	}
	if value := os.Getenv("OTEL_EXPORTER_OTLP_TIMEOUT"); value != "" {
		ms, err := strconv.ParseInt(value, 10, 64)
		if err != nil || ms <= 0 || ms > math.MaxInt64/int64(time.Millisecond) {
			return fmt.Errorf("OTEL_EXPORTER_OTLP_TIMEOUT must be positive milliseconds within time.Duration range")
		}
	}
	if value := os.Getenv("OTEL_EXPORTER_OTLP_COMPRESSION"); value != "" && value != "gzip" && value != "none" {
		return fmt.Errorf("OTEL_EXPORTER_OTLP_COMPRESSION must be gzip or none")
	}
	if value := os.Getenv("OTEL_EXPORTER_OTLP_INSECURE"); value != "" {
		if !strings.EqualFold(value, "true") && !strings.EqualFold(value, "false") || strings.EqualFold(value, "true") != (scheme == "http") {
			return fmt.Errorf("OTEL_EXPORTER_OTLP_INSECURE must agree with the endpoint's HTTP(S) scheme")
		}
	}
	if value := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"); value != "" {
		for _, pair := range strings.Split(value, ",") {
			name, encoded, ok := strings.Cut(pair, "=")
			value, err := url.PathUnescape(encoded)
			if !ok || err != nil || !httpguts.ValidHeaderFieldName(strings.TrimSpace(name)) || !httpguts.ValidHeaderFieldValue(value) {
				return fmt.Errorf("invalid OTEL_EXPORTER_OTLP_HEADERS")
			}
		}
	}
	if path := os.Getenv("OTEL_EXPORTER_OTLP_CERTIFICATE"); path != "" {
		pem, err := os.ReadFile(path)
		if scheme != "https" || err != nil || !x509.NewCertPool().AppendCertsFromPEM(pem) {
			return fmt.Errorf("OTEL_EXPORTER_OTLP_CERTIFICATE requires HTTPS and a readable PEM certificate")
		}
	}
	cert, key := os.Getenv("OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE"), os.Getenv("OTEL_EXPORTER_OTLP_CLIENT_KEY")
	if cert != "" || key != "" {
		if _, err := tls.LoadX509KeyPair(cert, key); err != nil || scheme != "https" {
			return fmt.Errorf("OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE and OTEL_EXPORTER_OTLP_CLIENT_KEY require HTTPS and a valid certificate/key pair")
		}
	}
	return nil
}
