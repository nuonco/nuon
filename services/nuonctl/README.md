# apictl

*****NOTE: This is not officially supported**

`apictl` is a small CLI which allows us to write queries and mutations in code to work more easily with our api.

## How it works

There are a set of query and mutation files in `queries` and `mutations` -- each query or mutation can be invoked using:

```bash
$ ./apictly.py mutation upsertOrg
```

Queries and mutations can be dynamic, for instance the following query:

```yaml
mutation UpsertUser($input: UserInput!) {
  upsertUser(input: $input) {
    id
    firstName
  }
}

## variables

input_envelope: true
firstName: Jon
lastName: morehouse
email: jon@nuon.co
```

Will create a mutation, and will pass all of the arguments in as `input:`. Furthermore, you can set the value of any argument in a query using an environment variable. For instance, if you change the above to:

```yaml
## variables

firstName: $FIRST_NAME
```

then the environment variable of `FIRST_NAME` will be used. This makes it useful to run queries and mutations against a specific org or user.


## Set up

You need to install `python3` and this application's dependencies:

```bash
brew install python3
pip install -r requirements.txt
```

Next, you need to set a token for each environment:

`APICTL_DEV_TOKEN` - graphql api token to use with dev
`APICTL_STAGE_TOKEN` - graphql api token to use with stage
`APICTL_PROD_TOKEN` - graphql api token to use with prod

If you'd like to lock onto a single environment, you can set `APICTL_ENV` to `dev` `stage` or `prod`.
