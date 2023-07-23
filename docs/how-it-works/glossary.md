# Glossary

This section provides definitions of key Nuon terms used in the documentation.

**Application -** this is how you represent your application in Nuon. You can create as many applications as you want.

**Artifact** - a component of your application that has been deployed to an ECR registry, and follows the OCI (Open Container Initiative) standard.

**Build Agent** - a piece of software, specific to a vendor, responsible for creating deployable artifacts. The build agent communicates with a vendor's dedicated Infrastructure Management Server (IMS), which tells the agent what to build. It is the only part of the system that has permission to push images to the vendor’s ECR repository.

**Component** - any part of your application's software or infrastructure. The components we currently support are those you can represent by a Docker image, Helm chart, or Terraform module.

**Customer** - an end-customer, who is a customer of Nuon's customer.

**ECR** - Elastic Container Registry, a managed service by AWS for deploying containerized application images and artifacts. Nuon uses this to store all components you configure.

**EKS** - Elastic Kubernetes Service, a managed Kubernetes service by AWS. Nuon uses this to create a Kubernetes cluster in the vendor's cloud account and in each customer's cloud account.

**Infrastructure Management Server (IMS)** - The server that powers a vendor’s build agent and the install agent in every customer account. The IMS is responsible for telling agents what to do, i.e., which images to sync, which components to deploy, etc.

**Install** - the configuration details for a customer account in which you will be installing components of your application. To create an install, you will need the customer to set up an AWS IAM account and provide you its Amazon Resource Number (ARN).

**Install Agent -** A piece of software, specific to a customer account, that manages the lifecycle of components in that customer's install. The install agent communicates with a vendor's dedicated IMS and runs jobs scheduled by it. It is the only part of the system that has permission to provision infrastructure in a customer’s cloud.

**Sandbox -** a set of base infrastructure elements, specific to a customer install, to run your app. It contains the following resources:

-   VPC with public/private subnets
-   EKS cluster
-   ECR registry
-   Certificate manager
-   External DNS

A sandbox ensures your application is isolated from other infrastructure in your customer's cloud. A sandbox also provides a consistent environment so if your application works in one customer’s account, it will work in all of them.

**OCI Registry** - a repository used to store and access OCI container images. OCI (Open Container Initiative) defines open standards for the storage, distribution, and execution of container images. Nuon uses ECR as its OCI registry.

**Organization** - this is how you represent your company in Nuon. Once you create an organization, you can add one or more applications.

**Vendor** - a Nuon customer that wants to deploy their SaaS application in a customer's cloud account. In the documentation, “you” refers to a vendor.

**VPC** - Virtual Private Cloud is a service that lets you launch AWS resources in a logically isolated virtual network that you define. Nuon creates a VPC inside the sandbox for each customer account to isolate your application from other resources in that customer's cloud.