# FAQs

**Will I need to refactor my code?**

There is no need to refactor your code, and you can deploy your application and infrastructure to customer clouds with just a few clicks. This means you can continue developing your application, while ensuring every customer has the latest version.

**What is the name of the infrastructure server we deploy?**

The server that controls the agents used for building and deploying components is called the Infrastructure Management Server (IMS). There is one IMS for each vendor, which lives in the vendor’s cloud account or Nuon’s dedicated account for the vendor.

**Why do we deploy a server per vendor?**

Each vendor has its own dedicated IMS for security and isolation. The IMS that controls a vendor’s installs and builds never serves another vendor. Also, the IMS is designed never to have direct access to a customer’s account

**How do we isolate your builds/infra from another vendor?**

Nuon is designed for isolation between vendors for all parts of the application. The IMS which manages installations, build agent, install agents, and ECR registry are never shared or accessible between vendors.

**What is a component?**

A component is a part of your application's software or infrastructure, represented as a Docker image, Helm chart, or Terraform module.

**What type of components do you support?**

Nuon currently supports three types of components: Docker images, Helm charts, or Terraform modules.

**How do I create components?**

You create components using the Component dialog in the Nuon UI.

**What is an install?**

An install represents a customer account in which you will be deploying and running components of your application. To create an install, you will need the customer to set up an AWS IAM account and provide you its Amazon Resource Number (ARN).

**How does an install work?**

An install works by creating a sandbox in the customer account, which contains an ECR registry and install agent specific to that customer. The install agent is responsible for provisioning, updating, and de-provisioning all infrastructure related to a component, under the direction of the IMS.

**What permissions are needed for an install?**

To create an install, you will need an ARN (Amazon Resource Name) for an IAM (Identity and Access Management) role. Each customer must create a new role in their AWS account, add two policies to it, and give you the ARN for the role. After the initial setup, the customer’s install becomes self-managing and cross-account IAM access is no longer needed.

**What resources does Nuon manage and what resources does a vendor manage?**

The vendor only has to manage the parts of their application, i.e., the Docker images, Helm charts, or Terraform modules that Nuon uses for creating and deploying components. Nuon manages all aspects of provisioning, updating, and de-provisioning the software and infrastructure resources in every customer account.

**How do I see the status of an install?**

To see the status of an install, click the tile for that install on the Overview or Install page.