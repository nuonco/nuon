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

# Shord IDs prefix reference

We use 26 characters long short IDs: a 3-character entity prefix, followed by a 23-character long nano ID. The current list of prefixes is the following:

| Prefix | Entity                                                |
| ------ | ----------------------------------------------------- |
| app    | App                                                   |
| art    | Artifact                                              |
| aws    | AWSSettings                                           |
| bld    | Builds                                                |
| cmp    | Component                                             |
| dpl    | Deployment                                            |
| dom    | Domain                                                |
| gcp    | GPCSettings                                           |
| inl    | Install                                               |
| ins    | Instance                                              |
| org    | Org                                                   |
| sec    | Secret                                                |
| snb    | Sandbox Versions                                      |
| usr    | UserOrgs                                              |
| def    | default prefix used in case none is provided as input |

# Timestamp Data Types and Representation

The timestamps we use in first-party application code flow through the following representation in our stack:

In postgres, the column type is `timestampz` ([docs](https://www.postgresql.org/docs/current/datatype-datetime.html#DATATYPE-DATETIME-INPUT)) which is an abbreviation postgres accepts for the official SQL data type `"timestamp with time zone"`. This is an 8-byte representation in postgres including date, time to 1 microsecond precision, and time zone. All of our times in postgres SHOULD be stored at UTC. (Pete confirmed this to be true as of 2023-05-24).

In golang using gorm, our structs use the go standard library `time.Time` struct.

At the gRPC/protobuf layer, we use the `google.type.DateTime` third party data type ([docs](https://pkg.go.dev/google.golang.org/genproto/googleapis/type/datetime)). This library is published by google [at github.com/googleapis/googleapis](https://github.com/googleapis/googleapis/blob/master/google/type/datetime.proto) and vendored by the buf schema registry at [buf.build/googleapis/googleapis](https://buf.build/googleapis/googleapis). The docs have a comment `// This type is more flexible than some applications may want.` and I suspect that may apply to our specific case. We may want to opt for a simpler type that has either a numeric or string representation at some point. But for the moment, this is what we use.

With our protoc golang code, the go struct for this is "google.golang.org/genproto/googleapis/type/datetime" `datetime.DateTime` and we have a conversion function. This function forces the time into UTC as a precaution in case a local timezone timestamp gets written into postgres somehow by mistake. Thus everything coming back in gRPC replies should be a UTC timestamp.

In the API Gateway, we use the buf-generated javascript implementation of `google.type.DateTime` which has the `google_type_datetime_pb.DateTime.toObject` function. We model this object in typescript with a custom type `TDateTimeObject` for static typing purposes. We convert this plain object into a [luxon DateTime](https://moment.github.io/luxon/api-docs/index.html#datetimefromobject) and then to a string with `.toISO()`. All ISO timestamp strings coming out of the API gateway currently should be in UTC and in this syntax: `2023-05-17T21:51:15.000Z`.


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

## How to: debug with delve and VS Code

Our golang services can be launched with the [delve](https://github.com/go-delve/delve) debugger, which runs a network service, and VS Code can attach "remotely" to that service to set breakpoints and step through the code.

Note that with the monorepo, debugging can be unusably slow if you launch VS Code from the monorepo root, so only open the directory containing the specific golang program you want to debug, for example `services/api` or `services/orgs-api` for best results.

Configure a VS Code launch configuration in `services/api/.vscode/launch.json` (or the directory for the service you are working with) similar to the following:

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
      "apiVersion": 1
    }
  ]
}
```

Then, in a terminal shell, launch it under the delve debugger. Here's an example for the `orgs-api`

```bash
cd services/orgs-api
nuonctl service exec --name=orgs-api -- dlv debug --headless --listen 127.0.0.1:2345 . -- server
```

Delve should get prepared and stop the program just prior to `main()` (I think). It will not start the program until a remote debug client attaches to the debugger process. So don't worry if you don't see any startup log output yet.

In VS Code, activate the "Run and Debug" activity from the activity bar (ctrl+shft+d). You should see a menu of launch configurations including the one we defined in our `launch.json` above. Choose that one and click the play button.

You should now be good to set breakpoints and do typical graphical debugger investigation tasks.

