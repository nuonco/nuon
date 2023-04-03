# API Gateway

Nuon's GraphQL API gateway

## Prerequisite

### Private npm packages

Before working with this project you'll need to create a personal access token for [GitHub](https://www.notion.so/nuon/Private-Github-Packages-8a51fd28b14f4673bd2b0848099ad811) & [Buf](https://www.notion.so/nuon/Private-Buf-Packages-4deda30863714c08bd08c0043d7cd42c). Once you've got these tokens you'll be able to download this project's private dependencies using `npm install`. To build & run this project as an OCI image refer to this doc on using [Earthly](https://www.notion.so/nuon/npm-6e3177b76f4e4df9ab9500b8d7fe5d28?pvs=4#8eb796e83c784931bc0b6d668bd6e060).

### .env config

You can copy the `.env.example` file & add the needed values `cp .env.example .env`.

**Auth 0**

You'll need to login the Nuon Auth0 account to get the `AUTH_ISSUER` & `AUTH_AUDIENCE` values. The `AUTH_AUDIENCE` can be found on the API Gateway under "applications/apis" & the `AUTH_ISSUER` can be found in the API Gateway settings domain under "applications/applications".

**GRPC Services**

Now you'll need to set the GRPC services you want to connect to the gateway. Each service can be connect by adding an object to the `SERVICES` array. (i.e. `SERVICES=[{ "name": "org", "url": "localhost:8080" }]`).

You can use the GRPC services on staging if you're connecting over Twingate.


## Usage
 
Now that you've installed the deps & configured the `.env` you can start the gateway with `npm start` or run it in dev mode using `npm run dev`. See [this guide](https://www.notion.so/nuon/Use-GraphQL-6a882d61c1a24e45bb5306158f6174f1) on how to authenticate & interact with the GQL gateway.

## Updating buf dependencies

When changes to our protobufs happen we'll need to manually update the dependencies for the gateway. To do this you'll need to `export` your Buf token in the terminal then run  `npm update` or `npm update {buf-package}`, this should update the `package-lock.json` file with the latest version of the grpc lib.

**Buf dependencies list**

* APIs `npm update @buf/nuon_apis.grpc_node`;
* Components `npm update @buf/nuon_components.grpc_node`;
* orgs-api `npm update @buf/nuon_orgs-api.grpc_node`;
* shared: `npm update @buf/nuon_shared.grpc_node`;

## Updating GQL types

Whenever the GQL schema changes you'll need to regenerate the types to Typescript. You can do this by running `npm run generate-gql-types` and committing the generated file.
