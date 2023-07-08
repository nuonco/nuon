# api

## interceptors

The interceptors package declares common interceptors for working with the api.

## gqlclient

A go client for interacting with our graphql-api. This uses the public endpoint, and can be used anywhere.

## client

A go client for interacting with our grpc-api. This must be used over twingate, or from within our VPC.
