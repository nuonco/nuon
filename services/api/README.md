# api
Repo for the Nuon gRPC API.

## Local setup

### Create a database first

To run the api locally, you need to have a `postgres` database running locally. You can start this using `docker-compose` by running:

```
$ docker-compose up
```

Docker-compose should create your `api` user and `api` database for you, but if it doesn't, within
the docker container run the following:

Once postgres is running, create `api` user and a database called `api`:

```bash
$ psql -c "CREATE USER api"
$ psql -U postgres -c "CREATE DATABASE api"
```

If you'd like to use a different database, you can configure environment variables to change them.

### Install temporal

We use [temporalite](https://github.com/temporalio/temporalite) to run a temporal server locally. While this won't actually process the jobs, it will enable us to emit workflow calls.

If you are using `docker-compose up` to manage dependencies, `temporalite` will automatically be started for you.

### Run earthly

We use earthly to run our builds in containers. After `docker-compose up` run `earthly +bin`.

### Install go and run the api

Next, install `go` - https://go.dev/doc/install - this api generally tracks the most recent stable release.

Once go is installed, grab dependencies, and run the application:

```bash
$ go mod download
$ go run . migrate up
$ go run . server
```

At this point, you should have the api running and accessible at `http://localhost:8080`.

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

Clone the [shared-configs repo](https://github.com/powertoolsdev/shared-configs) and then from your `api` repo run: `golangci-lint run -c ../shared-configs/golangci.yaml` (the command assumes the `shared-configs` repo has the same parent directory as `api`).

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

The protobufs for all the API's endpoints can be found in [the protos repo](https://github.com/powertoolsdev/protos/tree/main/api) and in [buf.build](https://buf.build/nuon/apis).



