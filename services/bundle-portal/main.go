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
	"os/user"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/airgap/day2state"
)

type config struct {
	state       string
	region      string
	profile     string
	addr        string
	requestedBy string
	allowedHost stringList
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
	flag.StringVar(&cfg.state, "state", "", "runner state directory or s3://bucket/prefix state URI")
	flag.StringVar(&cfg.region, "region", "", "AWS region for S3 state")
	flag.StringVar(&cfg.profile, "profile", "", "AWS profile for S3 state")
	flag.StringVar(&cfg.addr, "addr", "127.0.0.1:8080", "HTTP listen address")
	flag.StringVar(&cfg.requestedBy, "requested-by", "", "display name for portal dispatches")
	flag.Var(&cfg.allowedHost, "allowed-host", "accepted HTTP Host header for non-loopback deployment (repeatable)")
	flag.Parse()
	if cfg.state == "" {
		return fmt.Errorf("--state is required")
	}
	if cfg.requestedBy == "" {
		cfg.requestedBy = currentUsername()
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	store, err := day2state.New(ctx, cfg.state, cfg.profile, cfg.region)
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
	portal, err := newPortalServer(store, token, cfg.requestedBy, hosts, logger)
	if err != nil {
		return err
	}
	server := &http.Server{
		Handler:           portal.handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	logger.Info("bundle portal listening", zap.String("address", listener.Addr().String()), zap.String("state", cfg.state))
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

func currentUsername() string {
	current, err := user.Current()
	if err != nil {
		return ""
	}
	return current.Username
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
