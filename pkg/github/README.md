# Github

This package contains our wrappers and tooling for working with `github`.

## subpackages

### client

The client package contains our idiomatic client for working with the `github` api. This client supports loading credentials using a github installation ID.

```go
ghClient, err := client.New(g.v,
  client.WithAppID(appKeyID),
  client.WithAppKey(appKey),
)
if err != nil {
  return "", fmt.Errorf("unable to get github client: %w", err)
}

# use the client
resp, _, err := ghClient.CreateInstallationToken(ctx, g.InstallID, &github.InstallationTokenOptions{
  Repositories: []string{g.RepoName},
})
```

## todo + roadmap

* we would like to move the github tooling from `api` into this package, to make it a standard abstraction.
* we would like to move the `github` tooling from `workers-executors` to this package
