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
	"path/filepath"
	"strings"
	"syscall"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/runner/airgap/bundleupgrade"
	"github.com/nuonco/nuon/pkg/runner/airgap/day2state"
)

type config struct {
	state            string
	localState       string
	region           string
	profile          string
	addr             string
	requestedBy      string
	stackName        string
	allowedHost      stringList
	bundleArchive    stringList
	bucketPrefix     string
	deploymentID     string
	bundleKey        string
	bundleURI        string
	runnerImage      string
	ecrRegistry      string
	stackOutputsKey  string
	installInputsURI string
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
	flag.StringVar(&cfg.state, "state", "", "S3 URI for a real deployment's runner state")
	flag.StringVar(&cfg.localState, "local-state", "", "local runner state directory for development or replay")
	flag.StringVar(&cfg.region, "region", "", "AWS region for S3 state")
	flag.StringVar(&cfg.profile, "profile", "", "AWS profile for S3 state")
	flag.StringVar(&cfg.addr, "addr", "127.0.0.1:8080", "HTTP listen address")
	flag.StringVar(&cfg.requestedBy, "requested-by", "", "display name for portal dispatches")
	flag.StringVar(&cfg.stackName, "stack-name", "", "CloudFormation install stack for a real deployment")
	flag.Var(&cfg.allowedHost, "allowed-host", "accepted HTTP Host header for non-loopback deployment (repeatable)")
	flag.Var(&cfg.bundleArchive, "bundle-archive", "historical bundle archive used to reconstruct legacy action diffs (repeatable)")
	flag.StringVar(&cfg.bucketPrefix, "bucket-prefix", "", "deployment asset prefix for uploaded bundles")
	flag.StringVar(&cfg.deploymentID, "deployment-id", "", "deployment identity for uploaded bundles")
	flag.StringVar(&cfg.bundleKey, "bundle-key", "", "fixed S3 object key activated by the runner")
	flag.StringVar(&cfg.bundleURI, "bundle-uri", "", "fixed S3 bundle URI activated by the runner")
	flag.StringVar(&cfg.runnerImage, "runner-image", "", "runner image reference used by uploaded bundles")
	flag.StringVar(&cfg.ecrRegistry, "ecr-registry", "", "customer ECR registry used by uploaded bundles")
	flag.StringVar(&cfg.stackOutputsKey, "stack-outputs-key", "", "S3 object key containing install stack outputs")
	flag.StringVar(&cfg.installInputsURI, "install-inputs-uri", "", "optional S3 URI containing install inputs")
	flag.Parse()
	if err := cfg.validate(); err != nil {
		return err
	}
	if cfg.localState != "" {
		cfg.state = cfg.localState
	}
	if cfg.requestedBy == "" {
		cfg.requestedBy = currentUsername()
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	baseStore, err := day2state.New(ctx, cfg.state, cfg.profile, cfg.region)
	if err != nil {
		return err
	}
	runnerStore := day2state.WithPrefix(baseStore, day2state.RunnerNamespace)
	controlStore := day2state.WithPrefix(baseStore, day2state.ControlNamespace)
	store := day2state.ReadOverlay(runnerStore, controlStore, day2state.Legacy(baseStore))
	stackStore := store
	if parent, ok := stackOutputsRoot(cfg.state); ok {
		stackStore, err = day2state.New(ctx, parent, cfg.profile, cfg.region)
		if err != nil {
			return err
		}
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
	portal, err := newPortalServer(store, stackStore, token, cfg.requestedBy, hosts, logger)
	if err != nil {
		return err
	}
	portal.controlStore = controlStore
	portal.bundleActionDefinitions, err = loadBundleActionDefinitions(ctx, cfg.bundleArchive)
	if err != nil {
		return err
	}
	if cfg.stackName != "" {
		options := []func(*awsconfig.LoadOptions) error{}
		if cfg.profile != "" {
			options = append(options, awsconfig.WithSharedConfigProfile(cfg.profile))
		}
		if cfg.region != "" {
			options = append(options, awsconfig.WithRegion(cfg.region))
		}
		awsCfg, err := awsconfig.LoadDefaultConfig(ctx, options...)
		if err != nil {
			return fmt.Errorf("load AWS config for install stack: %w", err)
		}
		cfn := cloudformation.NewFromConfig(awsCfg)
		portal.installStackName = cfg.stackName
		portal.installStackReader = &awsInstallStackReader{client: cfn}
		portal.stackPlanner = cfn
		if cfg.bundleKey != "" && cfg.bundleURI != "" && cfg.runnerImage != "" {
			bucket, statePrefix, err := s3Location(cfg.state)
			if err != nil {
				return err
			}
			objects := s3.NewFromConfig(awsCfg)
			deployment := bundleupgrade.DeploymentContext{
				Bucket: bucket, StatePrefix: statePrefix,
				BucketPrefix: cfg.bucketPrefix, Region: awsCfg.Region,
				Image: cfg.runnerImage, ECRRegistry: cfg.ecrRegistry, DeploymentID: cfg.deploymentID,
				BundleKey: cfg.bundleKey, BundleURI: cfg.bundleURI, StackOutputsKey: cfg.stackOutputsKey, InstallInputsURI: cfg.installInputsURI,
			}
			portal.stageBundle = func(ctx context.Context, archivePath, archiveName string, progress func(bundleupgrade.Progress)) (*bundleupgrade.Result, error) {
				return bundleupgrade.Stage(ctx, bundleupgrade.Input{ArchivePath: archivePath, ArchiveName: archiveName, Deployment: deployment, Store: objects, Progress: progress})
			}
		}
	}
	server := &http.Server{
		Handler:           portal.handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Minute,
		WriteTimeout:      30 * time.Minute,
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

func s3Location(uri string) (string, string, error) {
	trimmed := strings.TrimPrefix(uri, "s3://")
	if trimmed == uri {
		return "", "", fmt.Errorf("expected s3 URI")
	}
	bucket, prefix, ok := strings.Cut(trimmed, "/")
	if !ok || bucket == "" {
		return "", "", fmt.Errorf("invalid s3 URI %q", uri)
	}
	return bucket, strings.Trim(prefix, "/") + "/", nil
}

func (c config) validate() error {
	if c.state == "" && c.localState == "" {
		return fmt.Errorf("real deployment context is required: pass --state and --stack-name, or use --local-state for development")
	}
	if c.state != "" && c.localState != "" {
		return fmt.Errorf("--state and --local-state cannot be used together")
	}
	if c.localState != "" {
		if c.stackName != "" {
			return fmt.Errorf("--stack-name cannot be used with --local-state")
		}
		return nil
	}
	if !strings.HasPrefix(c.state, "s3://") {
		return fmt.Errorf("--state must be an s3:// URI for a real deployment; use --local-state for a local directory")
	}
	if c.stackName == "" {
		return fmt.Errorf("--stack-name is required for a real deployment")
	}
	if c.bundleKey == "" || c.bundleURI == "" || c.runnerImage == "" || c.stackOutputsKey == "" {
		return fmt.Errorf("real deployment bundle uploads require --bundle-key, --bundle-uri, --runner-image, and --stack-outputs-key")
	}
	return nil
}

// stackOutputsRoot resolves the prefix holding stack-outputs/outputs.json,
// which the stack bootstrap writes as a sibling of the runner state prefix
// (<prefix>/state[/namespace] vs <prefix>/stack-outputs/). When the state
// location does not follow that layout, the portal reads stack outputs from
// the state store itself.
func stackOutputsRoot(state string) (string, bool) {
	trimmed := strings.TrimRight(state, "/")
	stateIndex := strings.LastIndex(trimmed, "/state")
	if stateIndex < 0 {
		return "", false
	}
	remainder := trimmed[stateIndex+len("/state"):]
	if remainder != "" && !strings.HasPrefix(remainder, "/") {
		return "", false
	}
	parent := trimmed[:stateIndex]
	if strings.HasPrefix(trimmed, "s3://") {
		if parent == "s3:/" {
			return "", false
		}
		return parent, true
	}
	if parent == "" {
		return string(filepath.Separator), true
	}
	return parent, true
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

func requestHostAllowed(hosts map[string]bool, hostport string) bool {
	hostport = requestHost(hostport)
	if hosts[hostport] {
		return true
	}
	host, _, err := net.SplitHostPort(hostport)
	return err == nil && hosts[host]
}
