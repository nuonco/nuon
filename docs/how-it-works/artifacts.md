# Artifacts

Each Nuon tenant creates + stores artifacts in an isolated ECR repository. At a high level, component artifacts are designed with the following principles:

* installs have no dependency on an external image
* all component artifacts are OCI compatible

### Artifact syncing

The first step of any deployment is creating an OCI artifact and syncing it to the correct installs before deployment.

All components sync images into the cloud account before deploying them for reliability and lineage purposes. Relying on an externally hosted container image is dangerous — the image can be deleted (causing downtime) or, even worse, updated to contain malicious software.

Each install agent is connected to a tenant’s infrastructure management server. The server directs each agent to synchronize the correct image into its environment. A one-time use token is created for the tenant’s ECR repository, directing the agent to pull the image and store it locally.

The agent running in the install is the **only** actor permitted to write to the install’s ECR repo. Once an artifact is written into the install’s ECR account, it is _**only**_ accessible by the agent and the nodes in the cluster.

### Container image artifacts

The following components result in a container image running in each customer’s install:

* public image
* private external image
* public build
* private build
* helm

The first step of deploying these components is to either create (using Docker) the OCI image from source or pull it from a public or private registry and store it in the tenant’s ECR repository.

These images are synced to the customer’s install before being deployed to prevent the tenant’s ECR repository from being required for an install’s uptime.

### Helm artifacts

Nuon supports two types of helm components:

* public, prebuilt helm charts
* helm charts from source

Currently, Nuon **does** allow deploying a public helm chart directly. This means that when a public helm chart is being deployed in your customer’s install, the helm chart will be accessed publicly (not synced). We have this on our roadmap currently to change.

Nuon supports building a helm chart from source — either in a public or private repository. The build agent will create an OCI artifact of the helm chart that is synced into the customer’s install before being deployed.

### Terraform artifacts

We are currently hard at work adding Terraform components. These components will work the same way other components work — by syncing the required binaries + source to execute the Terraform into the customer’s install as an OCI artifact.
