# nuon-runner-go

The SDK that powers the Nuon runner.

## Overview

The Nuon runner is deployed in each install, and powers everything from updates, releases and rollouts to day 2
monitoring.

Full documentation is available at https://runner.nuon.co/docs/index.html.

All endpoints in the API follow REST conventions and standard HTTP methods. You can find the OpenAPI Spec
[here](https://runner.nuon.co/oapi/v3)

## Installation

In your project, you can install the package directly using `go get`:

```bash
go get github.com/nuonco/nuon-runner-go
```

In your code, add the following import:

```
import nuonrunner "github.com/nuonco/nuon-runner-go"
```

## Create a client

Create a new api client, using an API key set in the environment.

```go
apiURL := "https://runner.nuon.co"
apiToken := os.Getenv("NUON_API_TOKEN")
runnerID := os.Getenv("NUON_RUNNER_ID")

apiClient, err := client.New(s.v,
  client.WithAuthToken(apiToken),
  client.WithURL(apiURL),
  client.WithRunnerID(orgID),
)
if err != nil {
  return fmt.Errorf("unable to get api client: %w", err)
}
```

## Example usage

### List current available jobs

```go
jobs, err := apiClient.AvailableJobs(ctx)
```

## Contributing

Please submit a PR, and if you would like help, contact us on our [community
slack](https://join.slack.com/t/nuoncommunity/shared_invite/zt-1q323vw9z-C8ztRP~HfWjZx6AXi50VRA).

You can generate mock code using:

```bash
$ go generate ./...
```

You can also change the open api spec to generate against, by setting the `API_URL` field to a different value:

```bash
$ NUON_API_URL=http://localhost:8081 go generate ./...
```
