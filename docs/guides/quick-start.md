# Quick start

## At a Glance

This guide will provide you with all the necessary information to use Nuon to run your software on your customers' cloud accounts. You will learn how to create organizations, add an application, configure components, and set up your first install.

## Create an organization

To get started with Nuon you’ll need to [sign in](https://app.nuon.co), you can use a google account to sign in. Once you’re signed in you’ll be prompted to create an organization for your company. You can name it whatever you’d like and can change it later if needed.

## Create an Application

Next, you will create your first application by giving it a name. You can create as many applications as your company requires, but we'll make one for now. After you name the application, we'll direct you to the application overview page.

## Configure Your Component

Now that you have created an organization and your first application, it's time to configure a component that you want to run on your customers' cloud accounts. You can do this from the application overview or components page. Click the "Add Component" button, enter the component's name, provide a URL to a public Docker container, and click "Add Component."

## Create Your First Install

You must create an install to run your application on your customers' cloud accounts. You can do this from the application overview or the installs page by clicking the "Add Install" button. Give your install a name, select an AWS region, enter the IAM ARN role for the AWS account, and click "Add Install" (learn more about IAM ARN for Installs [here](./iam-role-for-installs.md)).  Nuon will start provisioning the install, and once it is ready, you can start running your application.

## Trigger Component Deployment

Once Nuon provisions the install, you can trigger a deployment. Go to the components page and click on the component you created. When the panel opens, click the "Deploy Component" button to start the deployment process. Once the deployment is complete, you can view your running application at [need a link].
