# Mono

This is Nuon's mono repo. Currently it contains all of our go code.

## Getting started

You should be able to work with this repo just like any go repository. Before getting started, it's important to take note of the directory structure:

* `services` - each different directory in `services` represents a service. Each service contains it's own Earthfile,
  terraform, helm chart and more. Services are only deployed when their code changes, or when `pkg` changes.
* `pkg` - this is where all shared `go` code lives. Making changes in here can affect many (or all) services, so proceed
  with caution. When a pr is submitted with changes here, we build/test/lint all services and all of `pkg`.
* `pkg/types` - proto bufs are defined here and managed using [buf](https://buf.build/nuon/).
* `infra` - this is where terraform that is not tied to an individual service lives. Things like `orgs` will eventually
  live here as well.

# Environment setup

You need the following tools setup to work with this repo:

* Earthly
* Buf
* go

Once you have these setup, simply do the following to start working locally:

```bash
$ go mod download
$ go generate ./...
```

This will download dependencies, and generate all code needed to execute locally. From here, you can work with any pkg or service directly.
