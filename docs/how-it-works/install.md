# Install

Nuon enables you to run even the most complex products in millions of your customer's cloud accounts. When we set out to build Nuon, we designed every aspect of an install for isolation. Your customers' want your app in their cloud for security, reliability, and control purposes. We designed Nuon to allow you to ship and iterate on your product while giving them what they need.

Not only are each of your customer installs isolated from their _**own**_ infrastructure, but they are also isolated from the infrastructure powering your Nuon tenant. Your Nuon tenant never has direct access to an install and is not a point of failure for any installation. If your Nuon tenant is down, your customer installs **continue working**.

Furthermore, we designed installs to be self-managing. The install agent handles all provisioning, updating, and monitoring.

#### Overview

When you create a new install, your tenant's infrastructure does three things:

* provisions a sandbox in the customer's cloud
* provisions and configures an agent in the sandbox
* deploys all current components to the install

Installs were designed with security and reliability in mind:

* after setup, cross-account access is not a requirement to manage the installation
* when a tenant's infrastructure is down, the install keeps working
* all provisioning happens locally **inside the install**, and all artifacts deployed reside there.

#### Sandbox setup

Each Nuon install is sandboxed in your customer's cloud, which is isolated from other infrastructure in your customer's cloud. Furthermore, since a Nuon install provisions itself, it's even better to run it in a sub-account.

From a security and isolation perspective, isolating installs with a sandbox means less blast radius for your customer's infrastructure and makes integrating with other important parts of their systems easier.

When your customer provides an IAM role, your tenant's infrastructure management server does a one-time setup of the install. After the initial setup, your install is self-managing, and cross-account IAM access is no longer needed.

So what exactly is a sandbox? With Nuon, sandboxes are a full set of "base" infrastructure to run your app — depending on the sandbox, this will include:

* private VPC + public/private subnet
* hardened compute cluster
* locked down IAM roles for managing every aspect of the installation
* ECR registry for OCI artifacts deployed into the install

Sandboxes are designed to run in the most regulated environments and include best practices around security, network posture, and hardening of the runtime at every level.

Nuon ensures all of your sandboxes are up-to-date, hardened, and operating. Each sandbox provides a consistent environment, so if your application works in one, it will work in every.

#### AWS-EKS sandbox

The AWS-EKS sandbox provisions the following resources:

* AWS EKS cluster
* VPC with public/private subnets
* IAM role for the agent
* cert manager
* ECR repository for container artifacts
* DNS zone

#### AWS ECS sandbox

The agent that runs in each cloud account is orchestration-layer and cloud agnostic. While it supports different clouds, we plan to add ECS support next.

We are planning on launching an AWS ECS sandbox in Q2. Please reach out if you would like to know more.

#### Custom sandbox

Each sandbox is built using Terraform. Sandboxes are designed to be customizable, and eventually, we will enable both vendors and end-customers to be able to change sandboxes to add things like:

* custom networking
* custom container scanning
* custom Kubernetes version / compute runtime

Please reach out if you want to know more.

#### Install agent

Once an install sandbox is provisioned, the Nuon infrastructure management server provisions an agent which controls the install. This agent is responsible for managing the installation.

Because the install agent runs **inside** the install, cross—account access to your customer's cloud is never needed after the initial install. From a security posture perspective, this means your customer's install can be locked down and audited — every part of the agent is designed for isolation and security:

* every operation to create/update or destroy resources is treated as it's own job in the sandbox's compute runtime
* the agent **only** talks to the secure infrastructure management server
* every operation has a manifest

The agent uses a plugin system to interact with different component types — this means that the agent running in the account can "learn" about new types of deployments as your product evolves.

Every agent operation is auditable by the infrastructure management server — from jobs executed to container images being synced to any failure.

#### Components and artifacts

Once an install is provisioned, the Nuon agent automatically provisions all of the different parts of your application to get it running.

When a new install is created, your components' latest version is automatically deployed. Behind the scenes, each component is powered by OCI image artifacts, and each artifact is synced to your install before the agent deploys them. This means that an install _**never**_ has an external dependency on your tenant infrastructure.
