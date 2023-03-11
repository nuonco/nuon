# Nuon component protos

This package contains our protocol buffers for component configuration. Specifically, these protocol buffers are what the `workflows` layer would expect as inputs.

Most of the values in here should be exposed directly in our UI, as the workflows are designed to deterministically add the values needed.

## packages

Components contain the following packages, designed for configuring "stages".

### build

The build package contains protos for configuring a build. Currently, we are planning on supporting docker builds in public/private github repositories and "external images".

External images are used to pull images from a vendor's ECR repository in order to deploy them.

Behind the scenes, these components leverage the `docker` plugin in waypoint. [Link](https://developer.hashicorp.com/waypoint/plugins/docker).

### deploy

The deploy package contains protos for configuring deployments. There are two types of deployments we plan on initially supporting:

* container deployments - deploy a container into a customer account
* helm deployments - deploy a public or private helm chart (public first)

For helm deployments, we will support integrating with a VCS config for supporting private helm charts that live in a vendor's repository.

### configs

The configs package contains config protos for defining different types of configs. It's worth mentioning that these configs will be displayed differently in the UI based on other stages etc.

* env var configs - configs for a user specified environment variable. Not applicable for helm charts.
* helm values - configs for configuring a helm chart.

### vcs

The VCS package contains protos for VCS configuration. Currently we only support private and public github repositories.

VCS configs can be embedded in both build and deploy configs, (for instance to support a private helm chart).
