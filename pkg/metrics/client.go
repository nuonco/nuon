package metrics

import (
	"fmt"

	"github.com/DataDog/datadog-go/v5/statsd"
)

const (
	maxBytesPerPayload int = 4096
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_client.go -source=client.go -package=metrics
type dogstatsdClient interface {
	statsd.ClientInterface
}

// getClient returns a new dogstatsd client
func (w *writer) getClient() (dogstatsdClient, error) {
	w.clientonce.Do(func() {
		client, err := statsd.New(w.Address, statsd.WithMaxBytesPerPayload(maxBytesPerPayload))
		if err != nil {
			w.clienterr = fmt.Errorf("unable to get datadog client: %w", err)
			return
		}
		w.client = client
	})
	return w.client, w.clienterr
}
