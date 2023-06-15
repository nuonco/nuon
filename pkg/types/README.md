# protos

This directory is for managing protocol buffers at Nuon.

## repos

Each top level directory in this project corresponds to a buf repository. These generally break down into two categories:

### services

Repository services represent services that we expose, this means that their types should not be considered "portable" and shared between buf repos.

* api - the protos api
* workflows - our temporal workflow request/response params

**NOTE:** we should not create proto dependencies on services (eg: you should not import the api protos from another buf repo).

### objects

Objects represent core objects which can be shared throughout our system. They are not "owned" by a single service but rather shared amongst different protos and services.

* deployments - protos that contain things which live in the deployments buckets (plans etc)
* components - component definitions

**NOTE:** object protos are designed to be shared and are safe to import by services.

#### importing objects protos

To be able to import for example `components` protos to the `api` protos you need to do the following first in your local env:
- generate a [buf token](https://buf.build/settings/user)
- set the token as the value of the environment variable `BUF_TOKEN` (manually or as part of the `.netrc` file)

If these steps are not done first when you try to `go generate` you will get the error `Failure: repository "buf.build/nuon/components" was not found`.

## usage

Protos are generated during both CI and locally. They should no longer be checked into source control. To generate locally, run:
```bash
$ go generate ./...
```

Alternatively, if you have the `buf` cli present you can use `buf fmt` `buf lint` from within this folder.

### Buf registry and npm packages
We use the Buf registry to generate npm packages that are used by the api-gateway. Because we have several packages that reference each other and Buf creates a lock file for these references we can end up with broken npm packages. Best way to avoid this is whenever you make a change to a set of protos is to run `buf mod update` within each proto package in this order:

1. components
2. workflows
3. apis

This will guarantee that the generated npm packages are all correct.
