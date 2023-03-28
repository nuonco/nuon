package metrics

import (
	"fmt"

	"github.com/DataDog/datadog-go/v5/statsd"
)

const (
	maxBytesPerPayload int = 4096

	// we set the `HOST_IP` env var on all running pods in our cluster.
	dogstatsdHostEnvVar string = "HOST_IP"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_client.go -source=client.go -package=metrics
type dogstatsdClient interface {
	statsd.ClientInterface
}

// newDogstatsdClient returns a new dogstatsd client
func (w *writer) getClient() (dogstatsdClient, error) {
	if w.client != nil {
		return w.client, nil
	}

	client, err := statsd.New(w.Address, statsd.WithMaxBytesPerPayload(maxBytesPerPayload))
	if err != nil {
		return nil, fmt.Errorf("unable to get datadog client: %w", err)
	}

	w.client = client
	return client, nil
}
