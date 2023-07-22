# Sandboxes

Each install runs in an isolated sandbox inside the customer’s cloud. Nuon provisions an install in a sandbox for isolation and consistency. Sandboxes ensure:

* an install does not affect anything else running in the cloud account
* the install works the same everywhere

Sandboxing an application in the customer’s cloud ensures a tenant can create thousands or more installs, which will _**work**_ each time. Nuon hardens each sandbox and ensures that a sandbox is connected and provisioned correctly at all times.

#### Provisioning sandboxes

To create an install, all you (or your customer) need to do is create an IAM role that grants access for Nuon to the desired cloud account. From there, the first thing that Nuon will do is provision the sandbox for the install.

The initial provisioning of the sandbox is the **only** time that Nuon ever directly accesses a cloud account. All other provisioning happens from the agent running in the cloud account.

Under the hood, Nuon uses Terraform to provision sandboxes and maintains a hardened, secured, and optimized set for running in a customer’s cloud.

#### Types of sandboxes

Our infrastructure engine supports AWS, GCP, and Azure cloud environments. Based on customer feedback, we have started by offering **only** AWS sandboxes — if you or your customer has a different need, please let us know.

**aws-eks**

Nuon currently supports an AWS EKS sandbox. This sandbox provisions the following:

* VPC with public/private subnets
* EKS cluster
* cert manager
* external-dns

**aws-ecs**

We plan to offer an ECS sandbox that uses ECS Fargate for running software in early ‘23.

#### Bring your own sandbox.

We are building the ability for a customer to bring their own sandbox in Q2 ‘23. This means that a customer could bring their own VPC, compute cluster and other infrastructure.

While this is not generally recommended, as it prevents isolation between a customer’s and the application’s infrastructure — it is required for some environments. Your customer should opt to run a full install whenever possible and create the proper permissions between installs and other resources needed.
