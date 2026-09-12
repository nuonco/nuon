package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type config struct {
	controlPlaneURL string
	orgID           string
	installID       string
	apiToken        string
	brandName       string
	brandLogoURL    string
	brandFaviconURL string
	brandPrimary    string
	brandSupportURL string
	addr            string
	allowedHost     stringList
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer logger.Sync()
	if err := run(logger); err != nil {
		logger.Error("bundle portal stopped", zap.Error(err))
		os.Exit(1)
	}
}

func run(logger *zap.Logger) error {
	var cfg config
	flag.StringVar(&cfg.controlPlaneURL, "control-plane-url", "", "ctl-api base URL")
	flag.StringVar(&cfg.orgID, "org-id", "", "Nuon organization ID")
	flag.StringVar(&cfg.installID, "install-id", "", "customer-managed install ID")
	flag.StringVar(&cfg.apiToken, "api-token", "", "Nuon service-account API token")
	flag.StringVar(&cfg.brandName, "brand-name", "Deployment portal", "vendor product name shown in the portal")
	flag.StringVar(&cfg.brandLogoURL, "brand-logo-url", "", "same-origin or HTTPS vendor logo URL")
	flag.StringVar(&cfg.brandFaviconURL, "brand-favicon-url", "", "same-origin or HTTPS vendor favicon URL")
	flag.StringVar(&cfg.brandPrimary, "brand-primary-color", "#8040bf", "vendor primary color as a six-digit hex value")
	flag.StringVar(&cfg.brandSupportURL, "brand-support-url", "", "HTTPS vendor support URL")
	flag.StringVar(&cfg.addr, "addr", "127.0.0.1:8080", "HTTP listen address")
	flag.Var(&cfg.allowedHost, "allowed-host", "accepted HTTP Host header for non-loopback deployment (repeatable)")
	flag.Parse()
	if err := cfg.validate(); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runConnected(ctx, cfg, logger)
}

func (c config) validate() error {
	if err := c.branding().Validate(); err != nil {
		return err
	}
	for name, value := range map[string]string{
		"--control-plane-url": c.controlPlaneURL,
		"--org-id":            c.orgID,
		"--install-id":        c.installID,
		"--api-token":         c.apiToken,
	} {
		if value == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	_, err := newConnectedClient(c.controlPlaneURL, c.orgID, c.installID, c.apiToken)
	return err
}

func (c config) branding() portalBranding {
	branding := defaultPortalBranding()
	if c.brandName != "" {
		branding.Name = c.brandName
	}
	if c.brandPrimary != "" {
		branding.PrimaryColor = c.brandPrimary
	}
	branding.LogoURL = c.brandLogoURL
	branding.FaviconURL = c.brandFaviconURL
	branding.SupportURL = c.brandSupportURL
	return branding
}

func runConnected(ctx context.Context, cfg config, logger *zap.Logger) error {
	client, err := newConnectedClient(cfg.controlPlaneURL, cfg.orgID, cfg.installID, cfg.apiToken)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", cfg.addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.addr, err)
	}
	defer listener.Close()
	token, err := randomToken()
	if err != nil {
		return fmt.Errorf("generate CSRF token: %w", err)
	}
	hosts, err := allowedHosts(cfg.addr, listener.Addr(), cfg.allowedHost)
	if err != nil {
		return err
	}
	portal, err := newPortalServer(token, hosts, logger)
	if err != nil {
		return err
	}
	portal.branding = cfg.branding()
	portal.connected = client
	server := &http.Server{Handler: portal.handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Minute, WriteTimeout: 30 * time.Minute, IdleTimeout: 60 * time.Second}
	logger.Info("connected bundle portal listening", zap.String("address", listener.Addr().String()), zap.String("install_id", cfg.installID), zap.String("control_plane_url", cfg.controlPlaneURL))
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func randomToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

type stringList []string

func (s *stringList) String() string {
	return strings.Join(*s, ",")
}

func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func allowedHosts(configured string, actual net.Addr, extra []string) (map[string]bool, error) {
	hosts := map[string]bool{actual.String(): true}
	host, port, err := net.SplitHostPort(actual.String())
	if err != nil {
		return nil, fmt.Errorf("parse listen address %s: %w", actual.String(), err)
	}
	configuredHost, _, configuredErr := net.SplitHostPort(configured)
	if configuredErr == nil && configuredHost != "" {
		hosts[configuredHost] = true
		hosts[net.JoinHostPort(configuredHost, port)] = true
	}
	for _, value := range extra {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, fmt.Errorf("--allowed-host cannot be empty")
		}
		if _, _, err := net.SplitHostPort(value); err != nil {
			host := strings.Trim(value, "[]")
			hosts[host] = true
			value = net.JoinHostPort(host, port)
		}
		hosts[value] = true
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		hosts["localhost"] = true
		hosts["127.0.0.1"] = true
		hosts["[::1]"] = true
		hosts[net.JoinHostPort("localhost", port)] = true
		hosts[net.JoinHostPort("127.0.0.1", port)] = true
		hosts[net.JoinHostPort("::1", port)] = true
		return hosts, nil
	}
	if (configuredHost == "" || configuredHost == "0.0.0.0" || configuredHost == "::") && len(extra) == 0 {
		return nil, fmt.Errorf("--allowed-host is required when --addr binds a wildcard address")
	}
	return hosts, nil
}

func requestHost(hostport string) string {
	return strings.TrimSuffix(hostport, ".")
}

func requestHostAllowed(hosts map[string]bool, hostport string) bool {
	hostport = requestHost(hostport)
	if hosts[hostport] {
		return true
	}
	host, _, err := net.SplitHostPort(hostport)
	return err == nil && hosts[host]
}
