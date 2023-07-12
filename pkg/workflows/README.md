# Workflows

This package contains common packages for working with our workflows or handling side effects and outputs of these workflows.

## activities

Shared activities that all workflows can write.

## client

Standard client for triggering workflows.

## meta

Meta package for writing metadata into s3 from a workflow.

## worker

A standard workflow worker, that can be used to boot up a new worker using our standard internal configuration (logger, metrics, configs etc).

## dal

A client for reading/writing data from workflows, including:

* workflow requests and responses
* component outputs
* plans

The dal client is a sort of "catch all" dal for any data inputs/outputs to s3.

## Disclaimer

`artifacts` might not be the right place for our domain specific read/write tooling in the future - some of that data is inevitably changing with how it's being read/written and not everything is exclusive to workflows.

We are considering turning the waypoint server into a document store for most of this, and eventually that might replace `s3client` altogether.
