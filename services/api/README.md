# api

Repo for the Nuon gRPC API. Also known as the "core API" (`services/api`) to distinguish from the "orgs API" (`services/orgs-api`).

## How to: Setup for local development (docker & earthly stack)

This is onboarding setup only needed the first time you do development work on this service or when we make significant changes to our toolset.

* Set up a github account in the highly unlikely event that you do not have one
* Set up a [github personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token) and store it in your password manager
* Install [docker](https://www.docker.com/get-started/) **OR** [podman](https://podman.io/)
  * Throughout this documentation, we will document `docker` commands. If you prefer `podman`, adjust accordingly.
* Sign up for [buf.build](https://buf.build/)
* Create a buf token using the buf web interface
* Store your `BUF_TOKEN` in your password manager
* Store your `BUF_USER` username for use later
* Install [AWS SSO Util](https://github.com/benkehoe/aws-sso-util)
* Setup AWS SSO util

```bash
unset AWS_PROFILE
aws-sso-util configure populate
```

* Make sure that opens the correct browser window and you authenticate with the correct google nuon account
* Install [golang](https://go.dev/doc/install)
  * this api generally tracks the most recent stable release.

## How to: develop locally with docker-compose & earthly

* Do a fresh `aws-sso-util login` daily as needed
* Ensure your `BUF_TOKEN` is not expired and refresh as needed
* Setup your environment in your shell using secrets from your password manager

```bash
export AWS_PROFILE='stage.NuonPowerUser'
export AWS_REGION='us-west-2'
export EARTHLY_SECRETS="BUF_TOKEN=CHANGEME_PUT_YOUR_REAL_BUF_TOKEN_HERE,GITHUB_TOKEN=CHANGEME_PUT_YOUR_REAL_GITHUB_PERSONAL_ACCESS_TOKEN_HERE"
export EARTHLY_BUILD_ARGS='BUF_USER=CHANGEME_PUT_YOUR_REAL_BUF_USERNAME_HERE'
```
* Launch dependencies including postgresql and temporalite via `docker-compose`

```bash
# run from the monorepo root directory
cd mono
docker compose up -d
```

* Run the golang unit tests for this service in earthly

```bash
# run from services/api
cd services/api
earthly +test
```

## How to: develop locally with docker-compose & go

* Login to buf for protobuf builds: `buf registry login`
* Generate fresh golang code

```bash
# run from the monorepo root directory
cd mono
go generate -v ./...
```
* Start the dependencies

```
# run from the monorepo root directory
cd mono
docker compose up -d
```

* Set up your shell environment variables

```bash
export AWS_PROFILE='stage.NuonPowerUser'
export AWS_REGION='us-west-2'
export GIT_REF='HEAD'
export GITHUB_APP_KEY='CHANGEME_PUT_THE_FULL_SSH_KEY_FILE_HERE'
```

* Run the db migrations and start the golang server

```bash
# run from the services/api directory in the monorepo
cd services/api
go mod download
go run . migrate up
go run . server
```
At this point, you should have the api running and accessible at `http://localhost:8080`.

## Local setup

### Create a database first

To run the api locally, you need to have a `postgres` database running locally. You can start this using `docker-compose` by running:

```bash
# run from the monorepo root directory
cd mono
docker-compose up
```

Docker-compose should create your `api` user and `api` database for you, but if it doesn't, within
the docker container run the following:

Once postgres is running, create `api` user and a database called `api`:

```bash
psql -c "CREATE USER api"
psql -U postgres -c "CREATE DATABASE api"
```

If you'd like to use a different database, you can configure environment variables to change them.

### Install temporal

We use [temporalite](https://github.com/temporalio/temporalite) to run a temporal server locally. While this won't actually process the jobs, it will enable us to emit workflow calls.

If you are using `docker-compose up` to manage dependencies, `temporalite` will automatically be started for you.

## How to apply DB changes in stage
After your schema changes are merged, run this script:
```
function stage-migrate-api() {
  POD_ID=$(kubectl --context stage-nuon get pods -n default -o name -l app.kubernetes.io/name=api | head -n 1)
  echo "pod $POD_ID"
  kubectl exec --tty $POD_ID -- sh -c '/bin/service migrate up'
}
```
Remember to `aws-sso-util login` first.

## How to apply DB changes locally

1. Run `go run . migrate create FILENAME sql` where `FILENAME` is a descriptive name for your migration, e.g. `create_users_table`. This will create a placeholder file with that name under the `migrations` folder.
2. Edit the created file and add the SQL commands for your changes, e.g. `CREATE TABLE users...`
3. Run `go run . migrate up`

Alternatively you can set up the goose CLI and use it directly.
```
brew install goose
goose -s -dir "./migrations" create sandbox_versions_create sql
GOOSE_DRIVER=postgres GOOSE_DBSTRING="host=127.0.0.1 port=5432 user=api dbname=api sslmode=disable" goose -dir "./migrations" up
```

Use `go run . migrate status` to check the current status of DB migrations on your local DB.

## How to run tests locally

To run all tests: `go test -count=1 ./...` or specify which tests you want to run.

## How to run the linter locally

```bash
cd services/api
golangci-lint run
```

## How to send requests to the API

First, install grpcurl (cURL for hRPC).

Add these functions to your terminal's script file:

```bash
function nuon-api-stage-request () {
  cd ~/path/to/your/api/repo
  grpcurl -protoset <(buf build -o -) -plaintext -d "$1" api.nuon.us-west-2.stage.nuon.cloud:80 "$2"
  cd - > /dev/null
}

function nuon-api-local-request () {
  cd ~/path/to/your/api/repo
  grpcurl -protoset <(buf build -o -) -plaintext -d "$1" localhost:8080 "$2"
  cd - > /dev/null
}
```

Use them to send requests locally or in the staging environment (for staging you have to be connected with Twingate).

```bash
$ nuon-api-local-request '{"id":"369e9d9d-7fba-4c59-aa7b-4ab550d53be9"}' app.v1.AppsService/GetApp
$ nuon-api-stage-request '{"user_id":"233", "org_id": "123"}' user.v1.UsersService/UpsertOrgMember
```

The protobufs for all the API's endpoints can be found in the `pkg/types` directory in the monorepo and in [buf.build](https://buf.build/nuon/apis).
