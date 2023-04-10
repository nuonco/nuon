import { DataAwsIdentitystoreGroup } from "@cdktf/provider-aws/lib/data-aws-identitystore-group";
import { DataAwsSsoadminInstances } from "@cdktf/provider-aws/lib/data-aws-ssoadmin-instances";
import {
  AwsProvider,
  AwsProviderDefaultTags,
} from "@cdktf/provider-aws/lib/provider";
import { SsoadminAccountAssignment } from "@cdktf/provider-aws/lib/ssoadmin-account-assignment";
import { SsoadminManagedPolicyAttachment } from "@cdktf/provider-aws/lib/ssoadmin-managed-policy-attachment";
import { SsoadminPermissionSet } from "@cdktf/provider-aws/lib/ssoadmin-permission-set";
import { Fn, TerraformOutput, TerraformStack } from "cdktf";
import { Construct } from "constructs";
import { defaultRegion } from "./defaults";
import { TAWSAccount } from "./org";

interface ISSOConfig {
  accounts: TAWSAccount[];
  defaultTags: AwsProviderDefaultTags[];
}

// SSO creates and configures single sign on via GSuite
// See README.md for additional details and setup required
export class SSO extends TerraformStack {
  constructor(scope: Construct, name: string, config: ISSOConfig) {
    super(scope, name);

    const desiredRoles = [
      {
        description:
          "NuonAdmin provides administrator access within an account",
        groupName: "engineers-root",
        managedPolicyArns: [
          "arn:aws:iam::aws:policy/AdministratorAccess",
          "arn:aws:iam::aws:policy/AWSBillingReadOnlyAccess",
        ],
        name: "NuonAdmin",
        sessionDuration: "PT12H",
      },
      {
        description:
          "NuonPowerUser provides full access to AWS services and resources, but does not allow management of Users and groups.",
        groupName: "engineers",
        managedPolicyArns: [
          "arn:aws:iam::aws:policy/AdministratorAccess",
          "arn:aws:iam::aws:policy/AWSBillingReadOnlyAccess",
        ],
        name: "NuonPowerUser",
        sessionDuration: "PT12H",
      },
    ];

    new AwsProvider(this, "sso", {
      defaultTags: config.defaultTags,
      region: defaultRegion,
    });

    const ssoAdminInstance = new DataAwsSsoadminInstances(this, "this", {});

    const ssoInstanceARN = Fn.element(ssoAdminInstance.arns, 0);
    const ssoIdentityStoreId = Fn.element(ssoAdminInstance.identityStoreIds, 0);

    // for each role
    desiredRoles
      .map((role) => {
        return {
          ...role,
          // grab the group
          group: new DataAwsIdentitystoreGroup(this, `${role.name}-group`, {
            alternateIdentifier: {
              uniqueAttribute: {
                attributePath: "DisplayName",
                attributeValue: role.groupName,
              },
            },
            identityStoreId: ssoIdentityStoreId,
          }),
          // create the permission set
          permissionSet: new SsoadminPermissionSet(
            this,
            role.name.toLowerCase(),
            {
              description: role.description,
              instanceArn: ssoInstanceARN,
              name: role.name,
              sessionDuration: role.sessionDuration,
            }
          ),
        };
        // now that we have the necessary information
      })
      .forEach((role) => {
        // attach the desired managed policies
        role.managedPolicyArns.forEach((arn) => {
          const policyName = arn.split("/")[1];
          const id = `${role.name}-${policyName}`.toLowerCase();

          new SsoadminManagedPolicyAttachment(this, id, {
            instanceArn: ssoInstanceARN,
            managedPolicyArn: arn,
            permissionSetArn: role.permissionSet.arn,
          });
        });

        // and for each account and role,
        // map the group to the role and allow access to account
        config.accounts.forEach((acct) => {
          const id = `${role.name}-${acct.name}`.toLowerCase();
          new SsoadminAccountAssignment(this, id, {
            instanceArn: ssoInstanceARN,
            permissionSetArn: role.permissionSet.arn,
            principalId: role.group.id,
            principalType: "GROUP",
            targetId: acct.account.id,
            targetType: "AWS_ACCOUNT",
          });
        });
      });

    new TerraformOutput(this, "sso-last-updated", {
      description:
        "The timestamp of when the module was last applied. Useful for forcing applies on upgrade",
      value: new Date().toISOString(),
    });
  }
}
