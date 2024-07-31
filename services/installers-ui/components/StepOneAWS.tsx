import React, { type FC } from "react";
import { Link, Card } from "@/components";

export const StepOneAWS: FC<{ app: Record<string, any> }> = ({ app }) => {
  return (
    <div className="flex flex-col gap-4">
      <p>
        In order to connect your AWS account, please create an IAM role with
        permissions to provision your install. You can create your IAM role
        however works for your team, using one of the following mechanisms.
      </p>

      <p>
        This IAM role is only used when first provisioning your install, or when
        deprovisioning and can be removed after setup. Read more about{" "}
        <Link
          href="https://docs.nuon.co/guides/install-access-permissions"
          target="_blank"
          rel="noreferrer"
        >
          install access using AWS IAM
        </Link>{" "}
        in our documentation.
      </p>

      <Card className="p-4">
        <h3 className="text-md font-semibold">
          Create IAM policies with CloudFormation
        </h3>
        <p className="text-sm">
          You can create a 1-click IAM role with the correct policies attached
          to provision + deprovision your application using the following link.
          This will create an IAM role granting access to install{" "}
          {app?.metadata?.name}. Please use the stack output called{" "}
          <code className="bg-gray-800 text-orange-400 font-monospaced px-2 py-1 rounded-sm h-[22px] inline-flex text-[11px] tracking-wide">
            RoleARN
          </code>{" "}
          as an input in step 2.
        </p>

        <Link
          className="text-sm mt-4"
          href={`https://us-west-2.console.aws.amazon.com/cloudformation/home#/stacks/quickcreate?templateUrl=${app?.sandbox_config?.artifacts?.cloudformation_stack_template}&stackName=nuon-${app?.sandbox_config?.public_git_vcs_config?.directory}-permissions`}
          target="_blank"
          rel="noreferrer"
        >
          Create IAM Role
        </Link>
      </Card>

      <Card className="p-4">
        <h3 className="text-md font-semibold">
          Create IAM policies with Terraform
        </h3>
        <p className="text-sm">
          You can use our sandbox IAM role Terraform module to automatically
          create an IAM for a sandbox type using Terraform. Please set the IAM
          role ARN as an output, and use it for step 2.
        </p>

        <Link
          className="text-sm mt-4"
          href="https://github.com/nuonco/sandboxes/tree/main/iam-role"
          target="_blank"
          rel="noreferrer"
        >
          Terraform Module
        </Link>
      </Card>

      <Card className="p-4">
        <h3 className="text-md font-semibold">Create IAM policies manually</h3>
        <p className="text-sm">
          You can use the following policies to manually create an IAM role
          using the permissions noted below. Please note, the `deprovision`
          permissions are only needed if you would like to deprovision the
          install, and can be added later if desired.
        </p>

        <div className="flex gap-2 mt-4 items-center">
          <Link
            className="text-sm"
            href={app?.sandbox_config?.artifacts?.trust_policy}
            target="_blank"
            rel="noreferrer"
          >
            Trust policy
          </Link>
          <Link
            className="text-sm"
            href={app?.sandbox_config?.artifacts?.provision_policy}
            target="_blank"
            rel="noreferrer"
          >
            Provision policy
          </Link>
          <Link
            className="text-sm"
            href={app?.sandbox_config?.artifacts?.deprovision_policy}
            target="_blank"
            rel="noreferrer"
          >
            Deprovision policy
          </Link>
        </div>
      </Card>
    </div>
  );
};
