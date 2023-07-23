# Vendor Tenancy Model
Nuon provides a single-tenant environment with complete isolation between vendors for all parts of the application. The IMS which manages vendor installations, build agent, install agents, and ECR registry are never shared or accessible between tenants. See below for details on how each part of the system is isolated for security. 

## Infrastructure Management Server

Each vendor has its own dedicated IMS, which powers that vendor’s build agent and the install agents for all customers. The IMS is never shared between vendors. This means that the IMS powering a vendor’s installs never serves another vendor. Also, the IMS is designed never to have direct access to a customer’s account.

## Build Agent

Each vendor has its own build agent, responsible for performing all builds. This is the only part of the system that has access to a vendor’s private GitHub repository, ECR registry, or other sensitive information. The build agent communicates only with the vendor’s dedicated IMS for accessing jobs. This agent never performs builds for another vendor and is the only part of the system with permission to push images to the vendor’s ECR registry.

## ECR Registry for Vendor

The ECR registry for any vendor is only writable by that vendor’s build agent. Furthermore, all  artifacts deployed into a customer’s cloud are first synced to the vendor’s ECR registry. Syncs happen before deployments so the lineage of any container running in a customer’s account can be tracked and verified. This is important in some situations, e.g., when you deploy a public image and that image has been changed without you or your customer knowing.

## Sandbox

Each Nuon install is sandboxed in a customer's cloud, and isolated from all other infrastructure in that cloud. From a security perspective, isolating installs within a sandbox minimizes the impact on a customer's infrastructure, and makes integration with other systems easier. Sandboxes are designed to run in the most regulated environments and incorporate best practices around security, network posture, and hardening of the runtime at every level.

## ECR Registry for Customer

Each customer account has an ECR registry reserved for that customer only. The ECR registry associated with a customer's account never stores artifacts for any other customer.

Whenever a component is deployed, all OCI artifacts (container images, Helm charts, and Terraform modules) are synced from the vendor’s ECR registry to the install’s ECR registry. This serves two purposes:

-   **Security** - by syncing the images first into the install’s ECR registry, there is no need for the install agent to have long-lived access to the vendor’s ECR registry. In addition, the install agent can be restricted to only running images within the same account.
-   **Reliability** - Nuon is designed to support a vendor with thousands or millions of installs. Having all installs rely one vendor ECR registry ensures that the application components are deployed in exactly the same way.

The install agent only requires access to a vendor’s ECR registry when deploying a component. A one-time use token is created for the install to access the ECR registry and retrieve a specific artifact. If the vendor’s ECR registry is unavailable during the initial sync, the deployment will not complete. This ensures that an ECR outage for a vendor doesn’t create downtime for an install.

## Install Agent

Each customer account has its own install agent, responsible for managing the lifecycle of components in that install. This install agent runs locally in the customer account, and is never shared with any other customer’s account. It is the only part of the system that has permission to provision infrastructure in a customer’s cloud.

The install agent communicates with the dedicated IMS for its vendor and maintains a single outbound connection to it. Each job executes in its own container independently, and runs in the compute cluster provisioned by the sandbox.

Each vendor’s IMS server does a one-time setup of the install for any customer. After the initial setup, the customer’s install becomes self-managing, and cross-account IAM access is no longer needed.