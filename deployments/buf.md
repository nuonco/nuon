# Deployments protos

This package contains all internal protos for deployments (both builds and instance deploys) - eg: things like plans, outputs and other "state" we store in the `s3://nuon-org-deployments` bucket. As a reminder, the `nuon-org-deployments` bucket is namespaced by org shortID, eg: `s3://nuon-org-deployments/org=orgID/`.

## build

The build package contains protos that are emitted during a build. This includes the following:

* build plan - the plan which we use to execute the build
* build manifest - the output of the build (status, stats etc)
* build events - individual events that are emitted as part of a build

## deployments

The deployments package contains protos that are emitted during an instance deployment. This includes most of the things that we're pushing for deployments.
