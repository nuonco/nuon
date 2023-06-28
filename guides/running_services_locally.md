# Running services locally

## Background

Most of our services require AWS permissions to do anything useful. For instance, `workers-orgs` provisions infrastructure in the `orgs` accounts, all of which requires IAM roles that we can assume.

As a rule, all IAM roles that we create for services/operations in the system can be assumed by the `support-role`. The `support-role` is assumable by engineers who have access to AWS.

`nuonctl` leverages the support role to build a development experience where most services can be run and executed locally.

## dependencies

All services require that you are running the dependencies in `docker-compose` at the root of the mono repo. As of now, this starts the following services:

* `postgres` - for the api database
* `temporal` - for workflows
* `temporal-ui` - ui at localhost:8233 for developing.

## basic environment

To do most useful things, you need a reasonable starting environment. This includes:

* `aws` cli
* sso profiles in `$HOME/.aws/config`
* twingate setup
* a set of default environment variables.

To set up your environment, we recommend:

```bash
$ nuonctl scripts exec init-aws
$ nuonctl scripts exec init-kube
```

And then, we recommend setting some defaults:

```bash
AWS_PROFILE=stage.NuonPowerUser
AWS_REGION=us-west-2

# github token with access for GHCR and Node
GITHUB_TOKEN='<github-token>'
NUON_NPM_GITHUB_TOKEN='<buf-token>'
NUON_BUF_TOKEN='<buf-token>'
```
## running local on machine

`nuonctl service run-local` is used to run a service locally. This means that you need to have the dependencies installed for that particular application, and the correct version of go or node.

You can run any service using `nuonctl service run-local --name=foo`.

By default, this command will do the following:

* assume the support role and export credentials for it, so your service runs as "support", and can access real
  resources.
* authenticates with the stage kube cluster and fetches the config map for your service
* reads `service.yml` and sets environment variables that are denoted as "local" in there.
* runs one or more of the commands defined in `local_cmds`

Running services with `run-local` is a tradeoff - it will be a faster dev experience, but will come at the cost of possible deviation between local and stage, as the service is not running in the same container.

## executing commands with a stage environment

`nuonctl service exec-stage` allows you to run commands with the same environment as you would have in stage. This is useful for doing things like running a test suite, migrating a database and more.

For instance, to run the `api-gateway` integration scripts:

```bash
$ nctl service exec-stage --name=api-gateway npm run test:integration
```

## running a service locally, within a container

You can run a service locally inside of a container. Generally, this will be slow between restarts as the container has to cache / rebuild things. Our images aren't generally optimized for this yet.

In some cases, this is helpful and you can do so using:

```bash
$ nctl service run-earthly --name=api-gateway
```

## other service commands

`nuonctl service` is a great way to find all of the commands available for working with a service.
