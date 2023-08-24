# Changelog 

## Aug 18, 2023

### Terraform Provider

We have built a [Terraform provider](https://developer.hashicorp.com/terraform/language/providers) that enables you to interact with Nuon’s APIs from the command line. You can now specify Docker, Helm, or Terraform components, in your Nuon app, as Terraform resources. The Terraform provider also simplifies operations, such as creating an IAM role in the customer account. For details, see: https://github.com/powertoolsdev/quickstart#readme. 

### CLI Tool

You can now create installs, configure, build and deploy components, as well as check the status of installs and components, using the `nuon` command line tool. Previously, these operations were only supported in our web app. The CLI tool integrates with our Terraform provider. This makes it faster and easier to configure, deploy, and update your app, automate tasks, and integrate with your existing deployment workflow.

## July 21, 2023

### Private Helm Charts

Previously, you could only use a Helm chart that was packaged and in a public repo. We’ve overhauled how Helm charts work with Nuon, and cut out the middle layer. This means you can integrate any Helm chart with Nuon in 3 easy steps:

1.  Connect your GitHub repo.
2.  Specify the repo and directory your Helm chart is in.
3.  Override any values you want to customize with secrets, component outputs, and defaults.

### Connected Components

Remember when you wanted to build a self-hosted version of your product and realized you couldn’t use anything in the AWS console? Remember trying to package your entire app as a Helm chart, only to realize Minio didn’t quite work for you or the Kubernetes version of something broke your app?

Now, you can integrate components together and run them natively in your customer’s cloud. Here are some examples.

-   add a database using Terraform, and talk to it from an app — all running in the same cloud account
-   configure a Kafka cluster using MSK
-   run Temporal using Helm, backed by an RDS instance in the same database with a few workers and an API running

As always, you can run apps like this with a single click of a button (or an IAM role) for your customers. No more hours-long calls to set up a self-hosted version of your product.

### More Configuration Options

You can now configure components using a wider range of sources, secrets, and outputs.

-   Install outputs — When you create an install, we automatically create a Kubernetes cluster, VPC, and subnets. You can configure your Terraform and Helm modules to use these and do things like provision a new database in a private subnet in your customer’s cloud account.
-   Component outputs — Your app is going to need the URL, password, username for connecting to a database. Just export it from your Terraform, and add it as a config on your Helm chart.
-   Secrets — You can use secrets to configure Helm charts. For example, you can enter a Datadog key your customer provides, and it just works.

### Nuon Terraform Provider

We’ve built a Terraform provider for managing your Nuon configuration as code. Now you can do all of the following using Terraform.

-   create multiple apps
-   manage the configuration of components
-   create and update installs

This means you can use Terraform to manage a Nuon app, that runs Terraform against your customer’s cloud account.

### Go SDK

We’ve built a Go SDK, so you can build apps on top of our platform from your own backend. For example, if you build a sign-up flow to install your app in your customer’s cloud on day 1, it just works! There’s a lot more you can do with this. If you’re interested in learning more, just reach out!

### Reworking Component Configuration

In our next sprint, we’re completely rewiring how component configuration works to make it much more powerful. What does this mean for you? You'll be able to:

-   use outputs from Terraform components to power Helm charts
-   have an easier configuration experience when setting up apps
-   use secrets in new ways, in your customer cloud installs

## July 14, 2023

### Terraform Support

Previously, you could only deploy and run Docker/ECR images and public Helm charts in your customer cloud installs. We have now added support for Terraform components. This means you can connect a repo with Terraform to provision any infrastructure resources in your customer’s cloud account e.g., databases, SQS queues, and S3 buckets, in just a few minutes.

To provision infrastructure using Terraform, connect its GitHub repo to Nuon and add a Terraform component to your application. When you deploy a new version of this component, or create an install, Nuon automatically executes the Terraform inside your customer’s cloud account.

All infrastructure is provisioned using plugins running in the end-cloud account. In other words a plugin is executing Terraform to provision your resources in the end account. This is done for security, reliability, and isolation purposes.

This works using our standard package, sync, and provision workflow:

-   when you click Build (or use our API), we automatically create an OCI package of everything needed to provision your Terraform resources
-   when you deploy to an install, we automatically sync the package into the target customer’s cloud account
-   the target cloud account runner unpacks the OCI package, and executes a plugin to provision your Terraform resources

### Improved Internal Tooling for Plugins

We've improved our plugin infrastructure runtime, and built a set of internal tools for developing new plugins more easily. What does this mean for you?

-   we'll ship new features even faster
-   we're one step closer to letting you build your own plugins that run in your installs