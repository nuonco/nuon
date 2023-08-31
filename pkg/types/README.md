# protos

This directory is for managing protocol buffers at Nuon.

We generate all protos from the same branch for both Node and Go apps, and no longer push protos into the `buf` registry. You should never have to be logged into `buf` to generate protos here.

## repos

Each top level directory in this project corresponds to a buf repository. These generally break down into two categories:

### services

Repository services represent services that we expose, this means that their types should not be considered "portable" and shared between buf repos.

* workflows - our temporal workflow request/response params

### objects

Objects represent core objects which can be shared throughout our system. They are not "owned" by a single service but rather shared amongst different protos and services.

* external - vendored protos from upstreams. (Note: we can't call this vendor, as it will conflict with go's vendor
  directory).
* shared - shared types
* components - component definitions

**NOTE:** object protos are designed to be shared and are safe to import by services.

## usage

Protos are generated during both CI and locally. They should no longer be checked into source control. To generate locally, run:
```bash
$ go generate ./...
```

Alternatively, if you have the `buf` cli present you can use `buf fmt` `buf lint` from within this folder.
