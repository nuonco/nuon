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

### Install go and run the api

Next, install `go` - https://go.dev/doc/install - this api generally tracks the most recent stable release.

Once go is installed, grab dependencies, and run the application:

```bash
$ go mod download
$ go run . migrate up
$ go run . api
```

At this point, you should have the api running and accessible at `http://localhost:8080`.
