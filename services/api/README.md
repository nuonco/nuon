# api

Repo for the Nuon gRPC API. Also known as the "core API" (`services/api`) to distinguish from the "orgs API" (`services/orgs-api`).

## Developing

Some guidelines on working with your local instance of the API.

### How to send requests to the gRPC API

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

### How to apply DB changes locally

1. Run `go run . migrate create FILENAME sql` where `FILENAME` is a descriptive name for your migration, e.g. `create_users_table`. This will create a placeholder file with that name under the `migrations` folder.
2. Edit the created file and add the SQL commands for your changes, e.g. `CREATE TABLE users...`
3. Run `go run . migrate up`

### How to create a DB schema migration

* Install `goose` with `brew install goose` or your preferred package manager
  * https://github.com/pressly/goose
* Generate a template migration script
 * `goose -s -dir "./migrations" create sandbox_versions_create sql`
* Edit the corresponding `./migrations/*create_sandbox_versions.sql` file
* When ready to test the migration, run
  * `GOOSE_DRIVER=postgres GOOSE_DBSTRING="host=127.0.0.1 port=5432 user=api dbname=api sslmode=disable" goose -dir "./migrations" up`
* run `go run . migrate status` to check the current status of DB migrations on your local DB.

### How to get a psql prompt for DB work (local docker)

If postgres is not already running via `docker compose`, start it from the mono root directory with `docker compose up -d`.

Run `docker exec --interactive --tty mono-postgres-1 psql --username=api`

You can now run SQL, DDL, or weird postgres syntax like `\d orgs` to describe the `orgs` table.

### How to get a psql prompt for DB work on AWS RDS

Set up this helper shell function.

```bash
nuon-api-rds-shell() {
  PROFILE="$1"
  DB_HOST=$(aws --profile "${PROFILE}" rds describe-db-instances | jq -r '.DBInstances[0].Endpoint.Address')
  export PGPASSWORD=$(aws --profile "${PROFILE}" \
    rds generate-db-auth-token \
    --hostname "${DB_HOST}" \
    --port 5432 \
    --username api)
  echo "connecting to $DB_HOST -- please make sure you are connected to twingate ..."
  psql -h $DB_HOST -U api api
}
```
* `aws-sso-util login` as needed
* `twingate start`
* `nuon-api-rds-shell 'stage.NuonPowerUser'`

### How to run tests locally

To run all tests: `go test -count=1 ./...` or specify which tests you want to run.

### How to run the linter locally

```bash
cd services/api
golangci-lint run
```


