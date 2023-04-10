# Mono

This is Nuon's mono repo. Currently it contains all of our go code plus the API Gateway code.

## Getting started and filesystem organization

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

* [Earthly](https://earthly.dev/)
* [Buf](https://buf.build)
* [go](https://go.dev/)
* [protoc-gen-go-grpc](https://grpc.io/docs/languages/go/quickstart/)
* `go install github.com/srikrsna/protoc-gen-gotag`

Once you have these setup, do the following to start working locally:

```bash
buf registry login
$ go mod download
$ go generate ./...
```

This will download dependencies, and generate all code needed to execute locally. From here, you can work with any pkg or service directly.

# Basic Development Workflow

Generally, use `earthly ls` to see targets defined for any given `Earthfile`. We typically have `+test`, `+lint`, `+deploy` but these vary depending on the needs of a given sevice/project.

As new code is pulled in from git, running `go generate -v ./...` from the root of the monorepo will be necessary periodically if the changes affect generated types. You may also need to restart your lsp (language server protocol) server if your IDE functionality gets confused about the local filesystem state.
