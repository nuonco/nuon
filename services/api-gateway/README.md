# API Gateway

Nuon's GraphQL API gateway

## Prerequisite

### Private npm packages

Before working with this project you'll need to create a personal access token for [GitHub](https://www.notion.so/nuon/Private-Github-Packages-8a51fd28b14f4673bd2b0848099ad811)

Once you've got these tokens you'll be able to download this project's private dependencies using `npm install`. To build & run this project as an OCI image refer to this doc on using [Earthly](https://www.notion.so/nuon/npm-6e3177b76f4e4df9ab9500b8d7fe5d28?pvs=4#8eb796e83c784931bc0b6d668bd6e060).

### .env config

You can copy the `.env.example` file & add the needed values `cp .env.example .env`.

You should update the following environment variables in the file:

- `AUTH_AUDIENCE` and `AUTH_ISSUER`. To get the values you will need to log in the Nuon Auth0 account.
  - The `AUTH_AUDIENCE` can be found under "Applications / APIs / API Gateway / Settings / Identifier". Currently this is `api.nuon.co`.
  - The `AUTH_ISSUER` can be found under "Applications / Applications / API Gateway / Settings / Domain" and should be prefixed with `https://` so `https://nuon.us.auth0.com`
- `SERVICES`: this is a list of the GRPC services you want to connect to the gateway. Each service can be connect by adding an object to the `SERVICES` array. (i.e. `SERVICES=[{ "name": "org", "url": "localhost:8080" }]`). Note that in order to use the GRPC services on staging you need to be connected to Twingate. Sample `SERVICES` value to use the GRPC staging on services:
  ```
  SERVICES=[{"name":"status","url":"api.nuon.us-west-2.stage.nuon.cloud:80"},{"name":"org","url":"api.nuon.us-west-2.stage.nuon.cloud:80"},{"name":"app","url":"api.nuon.us-west-2.stage.nuon.cloud:80"},{"name":"install","url":"api.nuon.us-west-2.stage.nuon.cloud:80"},{"name":"component","url":"api.nuon.us-west-2.stage.nuon.cloud:80"},{"name":"deployment","url":"api.nuon.us-west-2.stage.nuon.cloud:80"},{"name":"github","url":"api.nuon.us-west-2.stage.nuon.cloud:80"},{"name":"instanceStatus","url":"orgs-api.nuon.us-west-2.stage.nuon.cloud:80"},{"name":"orgStatus","url":"orgs-api.nuon.us-west-2.stage.nuon.cloud:80"},{"name":"installStatus","url":"orgs-api.nuon.us-west-2.stage.nuon.cloud:80"}, {"name":"instance","url":"api.nuon.us-west-2.stage.nuon.cloud:80"}]
  ```
- `NUON_NPM_GITHUB_TOKEN`: your github personal access token you generated earlier

Make sure you do not source this `.env` file. Upon startup the gateway will read this file and set the appropriate environment variables.

## Usage

Now that you've installed the deps & configured the `.env` you can start the gateway with `npm start` or run it in dev mode using `npm run dev`. See [this guide](https://www.notion.so/nuon/Use-GraphQL-6a882d61c1a24e45bb5306158f6174f1) on how to authenticate & interact with the GQL gateway.

## Updating buf dependencies

We no longer use the buf schema registry for dependencies. All dependency clients from protobufs are automatically generated into the `src/build` directory when you run `go generate ./...` in the root of the mono repo.

Please refer to the wiki entry on code generation for more.

## Updating GQL types

Whenever the GQL schema changes you'll need to regenerate the types to Typescript. You can do this by running `npm run generate-gql-types` and committing the generated file.

## Adding a resolver

To add a new resolver we usually have to: update the package that contains the new grpc service, add the GQL schema for the query or mutation, list the query or mutation to the index file, write the resolver code, and add unit tests.

1. Add the GQL schema for the new resolver. This might require to update a package first in order to get the latest protos. Let's say we want to add a new query to retrieve all components for an organization.

If this is a new GRPC service then we would have to add something like this to the file `src/gql-schema/component.graphql`:

```graphql
extend type Query {
  componentsByOrg(orgId: ID!, options: ConnectionOptions): ComponentConnection!
}
```

2. Next step is to list the new query or mutation to the index file of the entity. Following the same example, we would edit `src/gql-resolvers/component/index.ts` and list the new query like this:

```ts
export const componentResolvers = {
  Query: {
    componentsByOrg,
  },
};
```

3. Finally, we write the code for the resolver. Add a new `.ts` file under the entity, following the same name convention as your query or mutation, and call the corresponding grpc service. Example:

```ts
export const components: TResolverFn<
  QueryComponentsArgs,
  Query["components"]
> = (_, { orgId }, { clients }) =>
  new Promise((resolve, reject) => {
    if (clients.component) {
      const request = new GetComponentsByOrgRequest().setOrgId(orgId);

      clients.component.getComponentsByOrg(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          const { componentsList } = res.toObject();

          resolve({
            edges:
              componentsList?.map((component) => ({
                cursor: component?.id,
                node: formatComponent(component),
              })) || [],
            pageInfo: {
              endCursor: null,
              hasNextPage: false,
              hasPreviousPage: false,
              startCursor: null,
            },
            totalCount: componentsList?.length || 0,
          });
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
```

To find the names of the methods to call (`GetComponentsByOrgRequest`, `getComponentsByOrg`) you will have to manually check the `pb` files of the imported package. For example, `./src/build/components/component/v1/messages_pb`.

Test your new resolver with Altair and if all is ok continue with writing unit tests for it.
