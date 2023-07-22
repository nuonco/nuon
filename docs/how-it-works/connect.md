# Connect

We've built Nuon to work with your existing workflow and architecture. Our goal is to make it a seamless part of your engineering team.

When we set out to build Nuon, we had a few goals in mind:

* make sure vendors do not need to support two projects
* coexist with your existing tools
* ability to add new infrastructure to customer clouds seamlessly

With Nuon, you should be able to create a working version of your product that can run in customer clouds on day 0 or day 1000 — all without making architecture and infrastructure tradeoffs.

### Overview

When you sign up for Nuon, we provision infrastructure that to manage your apps and installs:

* infrastructure management server
* build agent
* internal registry

This infrastructure is a single tenant deployed in your AWS account for even further isolation. Your installations **only** interact with this infrastructure, meaning Nuon _**never**_ has direct access to your customer's AWS accounts.

Each application part is a component — a component is a part of your stack, such as a terraform module, a helm chart, or just a basic container. Nuon uses a server-agent model to perform builds and deploys.

Under the hood, components use plugins to power things like builds, deploys, and operations — and you can even write custom plugins.

### Infrastructure Management Server (IMS)

When you create a new tenant with Nuon, we provision an infrastructure management server to power your components and installs.

The infrastructure management server does three things:

* controls build agents
* controls install agents
* stores historical information

The infrastructure management server is isolated for each tenant. This means the server your build agent and installs talk to is **never** the same server powering another tenant. Furthermore, the infrastructure management server can automatically run in your AWS account.

### Build agent + private registry

We create a deployment whenever you push to Nuon (whether with git directly or manually). A deployment represents a single "version" of your component running on each installation.

When you create a deployment, we first create a "build" that can be deployed in your customer's clouds.

We perform builds using an agent to isolate builds from all other customers. This agent communicates with the infrastructure management server, never talks to another tenant's infrastructure management server, and never touches another tenant's code, builds, or artifacts.

The build agent can build any component:

* basic components
* terraform component
* helm components
* custom components

Upon a successful build, the build agent pushes one or more artifacts into the tenant's isolated OCI-compliant registry.

The build agent and registry are deployed alongside the infrastructure management server in your cloud account. Nuon is designed for isolation between tenants at every layer, meaning an infrastructure-management server, agent, or registry will never be shared or accessible between tenants.

**Note:** component artifacts are synced to your customer's cloud so that the images running in an install live in that install. More in link.

### Components

You can have as many components as you like. We designed components to be pluggable, so you do not have to change your software/infrastructure to support bespoke deployments.

#### Basic components

Basic components allow you to deploy prebuilt images or buildable Github repositories in your customer's cloud.

You can use public images (for example, an open-source image like **httpbin**) or privately built images with basic components. Private images currently support AWS ECR. You can use them by supplying an IAM role that grants your Nuon build agent the ability to pull from your internal OCI repository.

Any repository that has a Dockerfile can be built and deployed using Nuon. You can supply both a public and private repository for basic containers. Your tenant's build agent will clone the repository, build an artifact using Kaniko and sync it to your tenant's OCI repository.

Images from prebuilt images and repositories are stored in your tenant's OCI repository and then synced to your customer's cloud before deployment.

#### Helm components

Helm components allow you to deploy prebuilt and source helm charts in your customer's cloud.

Public prebuilt helm charts are accessible via a public helm URL. These are "prebuilt" helm charts that are accessible from anywhere.

Source helm charts are helm charts that are supplied via a GitHub repository. Your tenant's build agent will clone the source repository and generate an OCI artifact of the helm chart that is synced to each install before use.

Public prebuilt helm charts are _**not**_ synced into your customer's cloud — a deployment may rely on those. We recommend using a source helm chart for production so that the source chart will be synced to your customer's cloud before use.

#### Terraform components

Terraform components allow you to provision custom resources in your customer's cloud. You can supply either a public or private repository with terraform source files.

Terraform components allow you to provision databases, storage buckets, and anything else your application needs. When you update a terraform component, your tenant's build agent will create an OCI-compatible bundle that contains everything needed to provision the terraform resources in your customer's cloud.

We are currently hard at work shipping terraform support and plan to support the following:

* ability to pass outputs from terraform to other components
* ability to run Terraform _**inside**_ the customer's install

If you would like to learn more, [reach out](https://www.notion.so/Nuon-s-Customer-Cloud-Architecture-e85b49da5e3d476fa177c6b80071dc31).

#### Coming soon — custom components

We designed Nuon so that any custom functionality can be supported using plugins. So, instead of changing your architecture, you can build a plugin to tell Nuon how to build and deploy a part of your application.

We plan to open up custom plugins to customers in Q2 or Q3 — if you have an urgent need, please reach out.

### Coming soon — connect using terraform

The fastest and easiest way to configure your application using Nuon is by using our UI. It should take only a few minutes to get started and set up as many components as you would like.

We plan on publishing a terraform provider that will let teams with many components configure their application with infrastructure as code. If you would like to learn more — [please reach out.](https://www.notion.so/Nuon-s-Customer-Cloud-Architecture-e85b49da5e3d476fa177c6b80071dc31)
