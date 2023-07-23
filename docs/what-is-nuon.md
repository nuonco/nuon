# What is Nuon?

Nuon is a platform that allows B2B software companies to deploy and run their applications in their customers' cloud accounts, quickly and securely. Nuon makes it easy to run even the most complex application in hundreds or thousands of your customer's cloud accounts, with each account isolated from all others. This enables you to meet the security and reliability concerns of your customers, without making architecture and infrastructure tradeoffs.

We've built Nuon to work with your existing workflow and architecture. There is no need to refactor your application, and you can deploy your application and infrastructure to customer clouds with just a few clicks. This means you can continue developing your application, while ensuring every customer has the latest version.

Nuon works in three easy steps:

-   **Configure** - specify the details of components of your application, such as Docker images, Helm charts, or Terraform modules
-   **Install** - deploy your application's components to a customer’s cloud account
-   **Run** - add or update components, and monitor the health of your application in each customer’s account

Let’s explore these steps in more detail.

## Configure

When you sign up for Nuon, the first step is to configure your application. This means adding its components, and specifying their source and deployment details.

Every application comprises different parts such as a frontend, backend, and AWS resources. In Nuon, we call these parts components. A component can be any of the following:

-   **Docker image**
    -   built from a private or public GitHub repository
    -   from Docker Hub or your ECR instance
-   **Helm chart** - this can be in a private or public repository
-   **Terraform module** - you can use this to provision any required infrastructure for your application, such as a database or S3 bucket

Once you’ve provided the source of a component, you must specify how to deploy it. Nuon supports two types of deployment configurations:

-   **Kubernetes**: set the pod configuration directly
-   **Helm chart**: from a private or public repository

For detailed steps, see: [Add a Component](guides/add-component.md).

## Install

Once you’ve configured the components of your application, the next step is to deploy them in a customer account.

In Nuon, an install represents a customer account in which you will be deploying and running your application. To create an install, you need to provide a name (typically the name of the customer) and specify the details of that customer’s AWS account. Your customer needs to create an AWS IAM role that grants you one-time access to their AWS account. Once you’ve created an install, you can deploy components to it.

For detailed steps, see:

-   [Create IAM Role for Install](guides/iam-role-for-installs.md)
-   [Create an Install](guides/create-install.md)
-   [Deploy a Component](guides/deploy-component.md)

## Run

Once you’ve deployed all components of your application in a customer's cloud, you will need to address these tasks.

-   Update existing components
-   Add new components.
-   Monitor the status of the application
-   Review logs and metrics, if something goes wrong

This part of the application lifecycle is called Day 2. Nuon gives you the necessary tools to update, modify, and operate your application in each customer's cloud account. When you add new components to your application and deploy them, they are automatically deployed into each customer's install.




