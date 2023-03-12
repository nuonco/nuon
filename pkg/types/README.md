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

We currently do _not_ support generating code in CI. This means that in order for your changes to be applied, you must run the generate step yourself locally.

The preferred way of doing this, is using `go generate`

```bash
$ go generate ./...
```

Alternatively, if you have the `buf` cli present you can use `buf fmt` `buf lint` from within this folder.
