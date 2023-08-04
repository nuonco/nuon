# Deprecated

Packages that will be removed, and should not be used in new ways.

## Helm

The helm package does not meet our needs/interface to the go-sdk in a way that we can build on top of. We plan on rebuilding this package at some point.

## Fetch

Fetch should not be used, and instead we should use `pkg/aws/s3downloader`.
