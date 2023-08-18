# api

## interceptors

The interceptors package declares common interceptors for working with the api.

## gqlclient

A go client for interacting with our graphql-api. This uses the public endpoint, and can be used anywhere.

## client

A go client for interacting with our grpc-api. This must be used over twingate, or from within our VPC.

## orgsclient

A go client for interacting with the orgs-api. This must be used over twingate, or from within our VPC.

**NOTE:** we are generally trying to minimize our reliance on the orgs-api, and plan to deprecate it in favor of using the org's waypoint server. Use with caution.
