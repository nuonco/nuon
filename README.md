# Mono

This is Nuon's mono repo. Currently it contains all of our go code plus the API Gateway code.

## Getting started and filesystem organization

You should be able to work with this repo just like any go repository. Before getting started, it's important to take note of the directory structure:

* `services` - each different directory in `services` represents a service. Each service contains it's own Earthfile,
  terraform, helm chart and more. Services are only deployed when their code changes, or when `pkg` changes.
* `bins` - applications that are compiled to binaries and not deployed as services. This can include `waypoint` plugins,
  as well as local clis and scripts.
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

Note, currently `nuonctl` requires `temporal` to run. You can start `temporal` using `docker-compose up -d` in the root of the repo.

Some helpful `nuonctl` commands:

* print a deployment plan: `nuonctl deployments print-plan <plan-path>`
* long id to short-id `nuonctl general to-short-id --id=<uuid>`
* start a canary `nuonctl general provision-canary`
* run a service locally `nuonctl service run --name=workers-apps`

# Services

Each directory in `services` represents a service, and is standardized. Services are built as container based images and deployed to our Kubernetes clusters.

All services must expose the following `earthly` targets:

* `docker` - build an image that can be pushed to ECR (also used to run locally)
* `test` - run tests such as unit tests or integration tests
* `lint` - run any linters

Services all expose a helm chart, in the `k8s` subdirectory, and terraform in the `infra` subdirectory. Nuonctl has some helpful commands for working with services locally:

* `nuonctl service exec` - execute a command in a service's environment
* `nuonctl service env` - print a service's stage environment
* `nuonctl service run` - run a service locally

# Development Tasks How-Tos

## How to: debug with delve and VS Code

Configure a VS Code launch configuration in `.vscode/launch.json` similar to the following:

```json
{
    // Use IntelliSense to learn about possible attributes.
    // Hover to view descriptions of existing attributes.
    // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [
        {
            "name": "delve localhost:2345",
            "type": "go",
            "request": "attach",
            "mode": "remote",
            "port": 2345,
            "host": "localhost",
            "apiVersion": 1,
        }
    ]
}
```

Then, in a terminal shell, load whatever environment variables the given service or go project needs, and launch it under the delve debugger. Here's an example for the `orgs-api`

```bash
cd services/orgs-api
dlv debug --headless --listen 127.0.0.1:2345 . -- server
```
Delve should get prepared and stop the program just prior to `main()` (I think). It will not start the program until a remote debug client attaches to the debugger process. So don't worry if you don't see any startup log output yet.

In VS Code, activate the "Run and Debug" activity from the activity bar (ctrl+shft+d). You should see a menu of launch configurations including the one we defined in our `launch.json` above. Choose that one and click the play button.

You should now be good to set breakpoints and do typical graphical debugger investigation tasks.