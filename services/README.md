# services

These are the services that make up the Nuon platform.

## Setup

We try to maintain a consistent set of tools and a similar workflow for all services. These directions should set you up to work on any of the services in this repo. Refer to the service READMEs for any service-specific directions.

* Set up a github account
    * Create a [github personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token) and store it in your password manager.
* Install [AWS SSO Util](https://github.com/benkehoe/aws-sso-util)
    * Setup AWS SSO util.
    * Set up an AWS profile config at `~/.aws/config (not sure where we get this from.)
* Install [docker](https://www.docker.com/get-started/) **OR** [podman](https://podman.io/)
* Sign up for [buf.build](https://buf.build/)
    * Create a buf token using the buf web interface
    * Store your `BUF_TOKEN` in your password manager
    * Store your `BUF_USER` username for use later
* Install the nuonctl CLI tool.
    * From `/mono/bins/nuonctl`, run `go install .`.

Then start the local instances of Temporal and Postgres defined in the docker-compose file at the root of the repo.

```sh
docker-compose up -d
```

And then configure your local environment.

```sh
export AWS_DEFAULT_PROFILE=stage.NuonPowerUser
export AWS_REGION='us-west-2'
export EARTHLY_SECRETS="BUF_TOKEN=CHANGEME_PUT_YOUR_REAL_BUF_TOKEN_HERE,GITHUB_TOKEN=CHANGEME_PUT_YOUR_REAL_GITHUB_PERSONAL_ACCESS_TOKEN_HERE"
export EARTHLY_BUILD_ARGS='BUF_USER=CHANGEME_PUT_YOUR_REAL_BUF_USERNAME_HERE'
```

Now that all the dependencies are installed, configured & running, you should be able to run any service locally using nuonctl.

```sh
nuonctl service run --name api
```
