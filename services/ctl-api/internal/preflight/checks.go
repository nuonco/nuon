package preflight

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"time"

	clickhousecore "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/jackc/pgx/v5"

	internal "github.com/nuonco/nuon/services/ctl-api/internal"
)

const checkTimeout = 5 * time.Second

// Checks is the registry of all preflight check implementations.
var Checks = map[string]Check{
	"rds":        checkRDS,
	"clickhouse": checkClickhouse,
	"temporal":   checkTemporal,
	"nuon-auth":  checkNuonAuth,
	"auth0":      checkAuth0,
	"github":     checkGithub,
	"aws":        checkAWS,
}

func checkRDS(cfg *internal.Config) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return "", fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close(ctx)

	var version string
	if err := conn.QueryRow(ctx, "SELECT version()").Scan(&version); err != nil {
		return "", fmt.Errorf("query failed: %w", err)
	}

	// Extract short version (e.g. "PostgreSQL 15.4")
	short := version
	if len(short) > 40 {
		short = short[:40]
	}

	summary := FormatFieldSummary("db_host", fmt.Sprintf("%s:%s", cfg.DBHost, cfg.DBPort))
	return fmt.Sprintf("%s %s", short, summary), nil
}

func checkClickhouse(cfg *internal.Config) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()

	var tlsCfg *tls.Config
	if cfg.ClickhouseDBUseTLS {
		tlsCfg = &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
		}
	}

	opts := &clickhousecore.Options{
		Addr: []string{fmt.Sprintf("%s:%s", cfg.ClickhouseDBHost, cfg.ClickhouseDBPort)},
		Auth: clickhousecore.Auth{
			Database: cfg.ClickhouseDBName,
			Username: cfg.ClickhouseDBUser,
			Password: cfg.ClickhouseDBPassword,
		},
		TLS:         tlsCfg,
		DialTimeout: cfg.ClickhouseDBDialTimeout,
		ReadTimeout: cfg.ClickhouseDBReadTimeout,
	}

	conn, err := clickhousecore.Open(opts)
	if err != nil {
		return "", fmt.Errorf("open failed: %w", err)
	}
	defer conn.Close()

	if err := conn.Ping(ctx); err != nil {
		return "", fmt.Errorf("ping failed: %w", err)
	}

	var version string
	if err := conn.QueryRow(ctx, "SELECT version()").Scan(&version); err != nil {
		// Non-fatal: connection works even if version query fails.
		version = "unknown"
	}

	summary := FormatFieldSummary("clickhouse_db_host", fmt.Sprintf("%s:%s", cfg.ClickhouseDBHost, cfg.ClickhouseDBPort))
	return fmt.Sprintf("ClickHouse %s %s", version, summary), nil
}

func checkTemporal(cfg *internal.Config) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()

	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", cfg.TemporalHost)
	if err != nil {
		return "", fmt.Errorf("dial failed: %w", err)
	}
	conn.Close()

	summary := FormatFieldSummary("temporal_host", cfg.TemporalHost)
	return fmt.Sprintf("connected %s", summary), nil
}

func checkNuonAuth(cfg *internal.Config) (string, error) {
	if cfg.NuonAuthIssuerURL == "" {
		return "", fmt.Errorf("nuon_auth_issuer_url is not set")
	}

	// Validate provider type if set.
	if cfg.NuonAuthProviderType != "" {
		known := map[string]bool{"auth0": true, "oidc": true, "cognito": true}
		if !known[cfg.NuonAuthProviderType] {
			return "", fmt.Errorf("unknown nuon_auth_provider_type: %s", cfg.NuonAuthProviderType)
		}
	}

	// Fetch OIDC discovery document.
	url := cfg.NuonAuthIssuerURL + "/.well-known/openid-configuration"

	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("request build failed: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("OIDC discovery failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OIDC discovery returned %d", resp.StatusCode)
	}

	summary := FormatFieldSummary("nuon_auth_issuer_url", cfg.NuonAuthIssuerURL)
	return fmt.Sprintf("OIDC discovery OK %s", summary), nil
}

func checkAuth0(cfg *internal.Config) (string, error) {
	url := cfg.Auth0IssuerURL + "/.well-known/openid-configuration"

	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("request build failed: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("OIDC discovery failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OIDC discovery returned %d", resp.StatusCode)
	}

	summary := FormatFieldSummary("auth0_issuer_url", cfg.Auth0IssuerURL)
	return fmt.Sprintf("OIDC discovery OK %s", summary), nil
}

func checkGithub(cfg *internal.Config) (string, error) {
	block, _ := pem.Decode([]byte(cfg.GithubAppKey))
	if block == nil {
		return "", fmt.Errorf("github_app_key is not valid PEM")
	}

	// Try to parse as PKCS1 or PKCS8 private key.
	_, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		_, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("github_app_key is not a valid RSA private key: %w", err)
		}
	}

	return fmt.Sprintf("valid PEM key, app ID %s", cfg.GithubAppID), nil
}

func checkAWS(cfg *internal.Config) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()

	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-west-2"))
	if err != nil {
		return "", fmt.Errorf("unable to load AWS config: %w", err)
	}

	stsClient := sts.NewFromConfig(awsCfg)

	// Assume the management role.
	creds := stscreds.NewAssumeRoleProvider(stsClient, cfg.ManagementIAMRoleARN)
	awsCfg.Credentials = creds

	assumedSTS := sts.NewFromConfig(awsCfg)
	identity, err := assumedSTS.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", fmt.Errorf("sts:GetCallerIdentity failed: %w", err)
	}

	return fmt.Sprintf("sts:GetCallerIdentity OK (%s)", *identity.Arn), nil
}
