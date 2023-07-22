# IAM role for Installs

## At a glance

This guide will explain how to add an Install in your customers' AWS account using an [Amazon Resource Name](https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html) (ARN) for an [Identity and Access Management](https://aws.amazon.com/iam/) (IAM) policy. We'll cover how to get the IAM trust policy, have your customer add it to their AWS account, and create an Install with the provided IAM ARN. To better understand how Nuon leverages AWS IAM policies and which permissions are needed, read our permissions documentation.

## Getting an IAM policy

TKTK

## Adding the policy

Now that you have an IAM trust policy, you'll need to give it to your customer's AWS admin so they can create the IAM resource needed for Nuon to provision the Install and deploy your application. Once the customer has made the IAM resource, they will need to provide you the ARN for said IAM resource.

## Adding an Install

Once you've received your customer's IAM ARN, you can add the Install. Next, go to the Installs page and click the "Add an Install" button in the top right corner. You'll need to name the Install (typically the customer's name), select an AWS region, and add the IAM ARN that your customer provided. Click "Add Install," and the Install will start provisioning. Install provisioning can take 15 to 30 minutes, but once complete, you can deploy your application to the customers' Install.
