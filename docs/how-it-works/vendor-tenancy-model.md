# Vendor tenancy model

We have built Nuon to be single-tenant and have designed the platform itself so tenants can run all or part of it in their own cloud.

#### Infrastructure

When a tenant signs up for Nuon, we provision the following infrastructure, which is used to power components and installations:

* infrastructure management server
* build agent
* ECR repository
* IAM roles + policies

Nuon is designed for isolation between tenants for all parts of the application. The server which manages tenant installations, the agents creating artifacts for components, and the permissions locking down all infra and its state, are isolated between tenants.

#### Infrastructure management server

Each tenant has their own server, which controls all of the installs. This server is responsible for telling agents what to do (which images to sync, which components to deploy, etc.) and storing long-lived telemetry and job data for each tenant’s installations.

The infrastructure management server is _**never**_ shared between tenants — this means that the server powering a tenant’s installations _**never**_ serves another tenant’s installs. Since the server is responsible for telling each agent what to do, it must be isolated from other tenants. It is also designed never to have access to an install directly — all provisioning happens inside the install with its agent.

The infrastructure management server can be run in a tenant’s own cloud account.

#### Build agent

Each tenant has a build agent responsible for creating deployable OCI artifacts. Each tenant’s build agent is isolated, meaning that the agent building one tenant’s components is **never** used to build another tenant’s components.

The tenant’s infrastructure management server manages the build agent. The server tells the agent **what** to build. (This is the same for all agents for a tenant).

Since each build agent is deployed for a single tenant, we can lock down the tenant’s ECR repository to **only** be writable by the tenant’s build agent. Furthermore, when a build needs access to a private Github repository, ECR repository, or otherwise sensitive system, **only** the agent accesses it.

The build agent can be run in a tenant’s cloud account.

#### ECR Repository

Each tenant has an isolated ECR repository reserved for that tenant **only**. The repository powering a tenant’s component artifacts _**never**_ stores artifacts for another tenant.

The ECR repository is **only** writable by a tenant’s agent. Furthermore, _**all**_ software artifacts deployed into a customer’s cloud are **first** synchronized to the tenant’s ECR repository — even public artifacts. Syncs happen before deployments so that the lineage of any container running in a customer’s account can be tracked and verified (for instance, if you deploy a public image, that image can be changed without you or your customer knowing).

All OCI artifacts (terraform modules, helm charts, and container images) are synchronized to the install’s ECR repository whenever a component is deployed. The two reasons are:

* Permissions — by syncing the images **first** into the install’s ECR repository, there is no need for long-lived access to the tenant’s ECR repo by an install. Furthermore, the install can be locked down to **only** allow running images in the same account.
* Reliability — Nuon is designed so a tenant can have thousands or millions of installs. If the install relies directly on a single tenant ECR repository, that means whenever a container is scheduled, the ECR repository must be active; otherwise, the install will not continue working.

During deployments is the **only** time that an install needs access to a tenant’s ECR repository. If the ECR repository is unavailable for the initial sync, the deployment will not go out — meaning that an ECR outage for a tenant will not create downtime for an install. A one-time use token is created for the install to access the ECR repository and “pull” a specific artifact.

The ECR repo for a tenant can be provisioned in a tenant’s own cloud.

### Nuon on your cloud

Today you can manually run most parts of Nuon in your cloud account. This means our team will help you install an infra management server, setup an agent, and provide the correct permissions.

Part of our Q2 roadmap is running Nuon in your own cloud — anyone can connect their own cloud to run Nuon on their own in minutes.
