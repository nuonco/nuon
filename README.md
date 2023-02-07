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

## Apply DB changes locally

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
