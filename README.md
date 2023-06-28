# Mono

This is Nuon's mono repo. Currently it contains all of our go code plus the API Gateway code.

## Getting started and filesystem organization

You should be able to work with this repo just like any go repository. Before getting started, it's important to take note of the directory structure:

- `services` - each different directory in `services` represents a service. Each service contains it's own Earthfile,
  terraform, helm chart and more. Services are only deployed when their code changes, or when `pkg` changes.
- `bins` - applications that are compiled to binaries and not deployed as services. This can include `waypoint` plugins,
  as well as local clis and scripts.
- `pkg` - this is where all shared `go` code lives. Making changes in here can affect many (or all) services, so proceed
  with caution. When a pr is submitted with changes here, we build/test/lint all services and all of `pkg`.
- `pkg/types` - proto bufs are defined here and managed using [buf](https://buf.build/nuon/).
- `infra` - this is where terraform that is not tied to an individual service lives. Things like `orgs` will eventually
  live here as well.

# Environment setup

You need the following tools setup to work with this repo:

- [Earthly](https://earthly.dev/)
- [Buf](https://buf.build)
- [go](https://go.dev/)
- [protoc-gen-go-grpc](https://grpc.io/docs/languages/go/quickstart/)
- [twingate](https://www.twingate.com/)

Once you have these setup, do the following to start working locally:

```bash
buf registry login
$ go mod download
$ go generate ./...
```

This will download dependencies, and generate all code needed to execute locally. From here, you can work with any pkg or service directly.

# Nuonctl

`Nuonctl` provides tools for working with services locally, triggering and inspecting workflows and various adhoc tasks. Any automations should be built into `nuonctl` first.

To run `nuonctl`, build the project into your `$PATH`:

```bash
$ cd bins/nuonctl
$ go build ~/bin/nuonctl .
```

Some helpful `nuonctl` commands:

- print a deployment plan: `nuonctl deployments print-plan <plan-path>`
- long id to short-id `nuonctl general to-short-id --id=<uuid>`
- start a canary `nuonctl general provision-canary`
- run a service locally `nuonctl service run --name=workers-apps`

# Services

Each directory in `services` represents a service, and is standardized. Services are built as container based images and deployed to our Kubernetes clusters.

All services must expose the following `earthly` targets:

- `docker` - build an image that can be pushed to ECR (also used to run locally)
- `test` - run tests such as unit tests or integration tests
- `lint` - run any linters

Services all expose a helm chart, in the `k8s` subdirectory, and terraform in the `infra` subdirectory. Nuonctl has some helpful commands for working with services locally:

- `nuonctl service exec` - execute a command in a service's environment
- `nuonctl service env` - print a service's stage environment
- `nuonctl service run` - run a service locally

# Validation with declarative tagging

We use 2 main libraries to assist with basic structural validation of our data and inputs.

- For protocol buffers, we use a plugin called `protoc-gen-validate`
  - This adds functionality to the basic `protoc` protocol buffer compiler
  - We can declare validations directly in our `.proto` source files
  - The list of built-in validations we can use is [in validate.proto here](https://github.com/bufbuild/protoc-gen-validate/blob/main/validate/validate.proto)
  - The syntax can be tricky. Study existing examples in code.
  - Use `(cd pkg/types && buf lint && buf build)` locally to check your syntax
- In regular go source code, we use [go-playground validator v10](https://pkg.go.dev/github.com/go-playground/validator/v10#section-documentation)
  - These we can use in `.go` source files as golang struct tags
  - golang structs get a generated method `.Validate()` we can use to trigger the validation


# Development Tasks How-Tos

## How to: Run a service locally with nuonctl

- `aws-sso-util login` as needed, typically once at the start of each work day
- `export AWS_PROFILE='stage.NuonPowerUser'`
- If you need access to resources in our stage VPC, run `twingate start`
- `cd mono`
- `nuonctl service run-local --name=orgs-api`t

## How to: share an org on stage with the team

- `aws-sso-util login` as needed
- grab your org id from the ui (short or long, either works)
- `nuonctl orgs add-nuon-users --org-id <orgid>`

