# Run

We designed Nuon so that every one of your customers can run your application in their cloud.

We discussed how an install is created and works in the install section. In this section, we will dive deep into the ongoing lifecycle of each of your customer’s installs.

#### Overview

Each install has a dedicated agent — this agent powers the lifecycle of the install:

* provisioning, updating, and deleting components
* syncing container images
* monitoring and observability
* debug commands

The agent is designed to be isolated and lock down permissions to each install. Since the agent provisions, updates, and manages the install, no long-lived access to the install is **ever** needed by you or Nuon.

The infrastructure management server for your tenant controls the agent.

### Install Agent

When a customer install is created, two pieces of infrastructure are provisioned:

* sandbox — the set of underlying compute, network, and cloud primitives to power the install
* agent — a process responsible for provisioning, updating, and managing the software running in the install.

The agent maintains a single outbound connection to a tenant’s infrastructure management server. This agent looks for three types of “jobs” to complete:

* component provisioning, updating, and de-provisioning
* log requests
* debug commands

#### Component job lifecycle

The agent is responsible for provisioning, updating, and de-provisioning all of the infrastructure for a component.

Each job to provision, de-provision, or update a component is performed using a remote job in the sandbox cluster. The agent itself _**does not**_ provision resources but **dispatches** jobs locally to do so. The agent distributes work in this way for resiliency and reliability — for instance if a terraform component is being applied, it could take up resources that would affect other components. Each job the agent kicks off executes one or more plugins to provision components. Each component type can be updated and changed without modifying the agent or its job(s). Components are modeled as plugins for improved reliability and customization. A plugin can be created for custom resources to bridge any provisioning requirements.

### Component artifacts

Components are designed to **only** rely on images that live in the installation’s current account. Whenever an install is created, an isolated OCI container repository is created and _**all**_ artifacts that are deployed in the account are synced to it first.

Whenever a component deployment is created, the tenant’s build agent builds one or more artifacts in the tenant’s ECR repository. This repository contains OCI artifacts for container images, terraform components, and helm charts.

Once a component artifact is ready to be deployed to an install, the agent synchronizes the image into its local ECR repository. This ECR repository is **only** accessible by the agent and its managed nodes.
