# Orgs-API

This is a protobuf service that exposes infrastructure information for orgs and their child resources. It combines data from `waypoint` and the `nuon-org-installations-*` and `nuon-org-deployments-*` buckets. It is not meant for public consumption, but rather to power parts of the UI via the `api-gateway` and internal tools. To access this, you must be on `twingate`.

## Implementation details

Protos for this application can be found on [buf](https://buf.build/nuon/orgs-api) or in the [protos repo](https://github.com/powertoolsdev/protos). This API uses buf.build's GRPC server [connect-go](https://github.com/bufbuild/connect-go).

Behind the scenes, this api is designed to ultimately be "single-tenant" - as many s3/waypoint calls as possible are locked down to a specific org using the appropriate IAM roles managed in `workers-orgs`.

## Roadmap

### org infrastructure

An org has a server which manages each of it's different agents.

- [ ] show status of the org server
- [ ] show status of the org runner
- [ ] show recent jobs from org runner/server

### installs

Installs are managed with terraform sandboxes and each come with their own runner.

- [ ] show current status of an install
- [ ] show all runs for an install (including versions, statuses)
- [ ] list resources for an install
- [ ] show status of install's runner

### deployments

Deployments represent a build/pull stage and a deploy stage to many different installs.

- [ ] list all deployments
- [ ] build logs for a deployment
- [ ] aggregate instance status for a deployment
- [ ] logs for a specific instance provision
