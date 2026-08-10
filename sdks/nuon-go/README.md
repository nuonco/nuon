# Nuon Go SDK

The Nuon Go SDK provides typed access to the Nuon control-plane API from Go programs.

For detailed guidance on authentication, BYOC compatibility, pagination, error handling, and testing, see the [Go SDK guide](https://docs.nuon.co/sdks).

## Installation

Install the SDK from the Nuon monorepo module:

```bash
go get github.com/nuonco/nuon/sdks/nuon-go
```

The former standalone module at `github.com/nuonco/nuon-go` is deprecated.

## Create a client

`WithURL` is required. Nuon Cloud clients should use `https://api.nuon.co`; BYOC clients must use the URL of their deployed control plane.

```go
package main

import (
	"context"
	"fmt"
	"os"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func main() {
	apiClient, err := nuon.New(
		nuon.WithURL("https://api.nuon.co"),
		nuon.WithAuthToken(os.Getenv("NUON_API_TOKEN")),
		nuon.WithOrgID(os.Getenv("NUON_ORG_ID")),
	)
	if err != nil {
		panic(fmt.Errorf("create Nuon API client: %w", err))
	}

	apps, hasMore, err := apiClient.GetApps(
		context.Background(),
		&models.GetPaginatedQuery{Limit: 100},
	)
	if err != nil {
		panic(fmt.Errorf("list apps: %w", err))
	}

	fmt.Printf("apps: %d, more pages: %t\n", len(apps), hasMore)
}
```

See the current [`Client` interface](./client.go) for the complete method list and signatures. The generated request and response types are in [`models`](./models).

## Contributing

The generated client uses the in-tree OpenAPI specification at `services/ctl-api/docs/public/swagger.json`. After changing the public API, regenerate the SDK from this directory:

```bash
go generate ./...
```

If the specification does not exist, the generation script creates it from the current control-plane source before running `go-swagger` and regenerating the client mock.
