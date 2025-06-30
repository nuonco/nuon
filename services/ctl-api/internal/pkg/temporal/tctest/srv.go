package tctest

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/cockroachdb/errors"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

type DevTestEnv struct {
	ClientOptions client.Options
	Server        *testsuite.DevServer
}

func NewEnv(ctx context.Context) (*DevTestEnv, error) {
	clientOpts := new(client.Options)
	var err error
	clientOpts.HostPort, err = getFreeHostPort()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get free host port")
	}

	clientOpts.Logger = slog.New(slog.DiscardHandler)
	// clientOpts.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	// 	Level: slog.LevelWarn,
	// }))

	srv, err := testsuite.StartDevServer(ctx, testsuite.DevServerOptions{
		ClientOptions: clientOpts,
		LogLevel:      "error", // server can add spurious noise to tests
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to set up dev server")
	}

	return &DevTestEnv{
		ClientOptions: *clientOpts,
		Server:        srv,
	}, nil
}

type ClientBits struct {
	Client client.Client
	// TClient  temporalclient.Client
	// EVClient eventloop.Client
	Worker worker.Worker
}

// TODO(sdboyer) should we rely on dependency injection, at least to some degree?
func (e *DevTestEnv) NewRunInNamespace(t *testing.T, ctx context.Context, namespace string) (*ClientBits, error) {
	clientOpts := e.ClientOptions
	clientOpts.Namespace = namespace

	namespaceClient, err := client.NewNamespaceClient(clientOpts)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create namespace client")
	}

	retention := time.Hour
	err = namespaceClient.Register(context.Background(), &workflowservice.RegisterNamespaceRequest{
		Namespace:                        namespace,
		WorkflowExecutionRetentionPeriod: &retention,
	})
	if err != nil {
		return nil, err
	}

	namespaceClient.Close()

	c, err := client.NewClientFromExisting(e.Server.Client(), clientOpts)
	if err != nil {
		t.Fatal(errors.Wrap(err, "failed to create client"))
	}

	wrk := worker.New(c, "default", worker.Options{})
	if err := wrk.Start(); err != nil {
		t.Fatal(errors.Wrap(err, "failed to start worker"))
	}

	return &ClientBits{
		Client: c,
		Worker: wrk,
	}, nil
}
