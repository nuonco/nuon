# Terraform Provider Nuon

A terraform provider for interacting with Nuon.

## Developing

First, sign into staging. You'll need an api token from stage.

Then, get [ctl-api](../../services/ctl-api) running locally in sandbox mode, following the directions in the [ctl-api development doc](../../wiki/ctl_api.md) The acceptance tests need a "real" instance of the API to run against, but we also don't want to trigger Temporal an create real resources while developing.

Next, run [./scripts/create-org.sh](./scripts/create-org.sh) to create an org in your local DB.

Then you can set up your local env:

```sh
export NUON_API_TOKEN=${an api token from stage}
export NUON_ORG_ID=${your_org_id}
```

You should be able to run `go test ./...` in this directory. All the tests should pass, and you should see the provider making requests in the api logs.
