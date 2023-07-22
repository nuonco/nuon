# Server and agents

Nuon uses a server/agent model for managing customer cloud installs. Each tenant has an infrastructure management server that controls a set of agents that perform builds, deploys, and other parts of the component lifecycle.

The server/agent model provides isolation and allows for a smaller security perimeter at each part of an install and component’s lifecycle. In other words, the agent/server model means that cross-account IAM permissions are not required for managing an install after the initial setup. All operations to provision, update and manage components are executed by an agent running in the install.

### **Build agent**

The build agent runs in Nuon’s dedicated account for tenants — or you can run it in your own account.

The build agent communicates only with the dedicated infrastructure management server for accessing jobs. Like the install agent, each job is executed as its own container process, meaning many jobs can be executed simultaneously.

The build job process can only build a tenant’s components; the agent and jobs are never shared between tenants.

### **Install agent**

The install agent runs locally in the customer’s cloud account. The agent process itself only has access to provision jobs locally and communicates with the dedicated infrastructure management server.

Each job executes its own container independently. The jobs are the only part of the system with access to modify the customer’s cloud. These jobs run in the compute cluster provisioned by the sandbox.

### **Agent execution model**

Each install has an agent that runs locally in the account. This agent runs jobs scheduled by its tenant’s dedicated infrastructure management server.

We designed the agent to run every job as its own process, so the agent is never overloaded. This means many jobs can be performed in parallel as they run independently.

When a job is executed, it will use a plugin system for each type of component — components are provisioned by plugins that power functionality, such as Helm, Terraform, and more. This means new types of infrastructure can be added over time. Eventually, vendors can provide their own plugins for bespoke deployments.

### **Tenancy model**

Each part of the system is designed to be run in different environments and is never shared between tenants or installs.

Each tenant has their own infrastructure management server, which powers the tenant’s build agent and the agents in each install. This server is never shared between tenants and can run in a vendor’s cloud account.

Each tenant has their own build agent, which performs all builds. This agent never performs builds for another tenant and is the only system that has permissions to push images to the vendor’s ECR repository.

Each install has its own dedicated agent, which manages the lifecycle of components in the install. This agent is never shared with another install and is the only thing that is permitted to provision infrastructure in the customer’s cloud. The agent communicates with the dedicated infrastructure management server for it’s tenant and maintains a single outbound connection to it.
