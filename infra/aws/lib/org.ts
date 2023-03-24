import { OrganizationsAccount } from "@cdktf/provider-aws/lib/organizations-account";
import { OrganizationsOrganization } from "@cdktf/provider-aws/lib/organizations-organization";
import { OrganizationsOrganizationalUnit } from "@cdktf/provider-aws/lib/organizations-organizational-unit";
import { OrganizationsPolicy } from "@cdktf/provider-aws/lib/organizations-policy";
import { OrganizationsPolicyAttachment } from "@cdktf/provider-aws/lib/organizations-policy-attachment";
import {
  AwsProvider,
  AwsProviderDefaultTags,
} from "@cdktf/provider-aws/lib/provider";
import { TerraformOutput, TerraformStack } from "cdktf";
import { Construct } from "constructs";
// eslint-disable-next-line import/no-unresolved
import { OuScp } from "./../.gen/modules/trussworks/aws/ou-scp";
import * as tagServices from "./data/tag-enforcement-services.json";
import { allowedRegions, defaultRegion, desiredOrgStructure } from "./defaults";

interface IOrgConfig {
  defaultTags: AwsProviderDefaultTags[];
}

export type TAWSAccount = {
  account: OrganizationsAccount;
  name: string;
};

export class Org extends TerraformStack {
  public accounts: TAWSAccount[];
  public org: OrganizationsOrganization;

  constructor(scope: Construct, name: string, config: IOrgConfig) {
    super(scope, name);

    new AwsProvider(this, "mgmt", {
      defaultTags: config.defaultTags,
      region: defaultRegion,
    });

    this.org = new OrganizationsOrganization(this, "org", {
      awsServiceAccessPrincipals: [
        "access-analyzer.amazonaws.com",
        "account.amazonaws.com",
        "auditmanager.amazonaws.com",
        "aws-artifact-account-sync.amazonaws.com",
        "backup.amazonaws.com",
        "member.org.stacksets.cloudformation.amazonaws.com",
        "cloudtrail.amazonaws.com",
        "config.amazonaws.com",
        "fms.amazonaws.com",
        "guardduty.amazonaws.com",
        "health.amazonaws.com",
        "inspector2.amazonaws.com",
        "ram.amazonaws.com",
        "securityhub.amazonaws.com",
        "storage-lens.s3.amazonaws.com",
        "servicequotas.amazonaws.com",
        "sso.amazonaws.com",
        "ssm.amazonaws.com",
        "tagpolicies.tag.amazonaws.com",
        "reporting.trustedadvisor.amazonaws.com",
      ],
      enabledPolicyTypes: [
        "AISERVICES_OPT_OUT_POLICY",
        "BACKUP_POLICY",
        "SERVICE_CONTROL_POLICY",
        "TAG_POLICY",
      ],
      featureSet: "ALL",
    });

    const orgRoot = this.org.roots.get(0).id;

    // opt out of AWS using our data for AI training
    const aiOptOutPolicy = new OrganizationsPolicy(this, "ai_opt_out_policy", {
      content: JSON.stringify({
        services: { default: { opt_out_policy: { "@@assign": "optOut" } } },
      }),
      name: "ai-opt-out",
      type: "AISERVICES_OPT_OUT_POLICY",
    });

    const orgTagPolicy = new OrganizationsPolicy(this, "env_tag_policy", {
      content: JSON.stringify({
        tags: {
          environment: {
            enforced_for: { "@@assign": tagServices },
            tag_key: { "@@assign": "environment" },
            tag_value: { "@@assign": ["management", "shared"] },
          },
        },
      }),
      name: "org-environment-tag-policy",
      type: "TAG_POLICY",
    });

    new OrganizationsPolicyAttachment(this, "ai_opt_out_root", {
      policyId: aiOptOutPolicy.id,
      targetId: orgRoot,
    });

    new OrganizationsPolicyAttachment(this, "env_tag_policy_root", {
      policyId: orgTagPolicy.id,
      targetId: orgRoot,
    });

    // add the created ou to the structure
    const units = desiredOrgStructure.map((ou) => {
      return {
        ...ou,
        unit: new OrganizationsOrganizationalUnit(this, ou.name, {
          name: ou.name,
          parentId: orgRoot,
        }),
      };
    });

    // create SCPs for OUs
    units.forEach((ou) => {
      if (ou.disableScp) {
        return;
      }
      new OuScp(this, `github_terraform_aws_ou_scp-${ou.name}`, {
        allowedRegions,
        denyCreatingIamUsers: true,
        denyLeavingOrgs: true,
        denyRootAccount: true,
        denyS3BucketPublicAccessResources: ["arn:aws:s3:::*"],
        denyS3BucketsPublicAccess: true,
        limitRegions: true,
        protectIamRoleResources: [
          "arn:aws:iam::*:role/AWSReservedSSO*",
          "arn:aws:iam::*:role/AWSServiceRole*",
          "arn:aws:iam::*:role/OrganizationAccountAccessRole",
        ],
        protectIamRoles: true,
        target: {
          id: ou.unit.id,
          name: ou.unit.name,
        },
      });
    });

    // create and export the full set of accounts
    this.accounts = units.flatMap((ou) => {
      return ou.accounts.map((acct) => {
        const account = new OrganizationsAccount(this, acct.name, {
          closeOnDeletion: true,
          email: `aws-accounts+${acct.name}@powertools.dev`,
          iamUserAccessToBilling: "ALLOW",
          // this is a bit of a hack as this doesn't import
          // and on future plan's wants to recreate :grimacing:
          lifecycle: { ignoreChanges: ["iam_user_access_to_billing"] },
          name: acct.name,
          parentId: ou.unit.id,
        });

        const tagsToAdd = acct.additionalEnvTags || [];
        const acctEnvTagPolicy = new OrganizationsPolicy(
          this,
          `env_tag_policy_${acct.name}`,
          {
            content: JSON.stringify({
              tags: {
                environment: {
                  tag_value: { "@@append": [acct.name, ...tagsToAdd] },
                },
              },
            }),
            name: `${acct.name}-environment-tag-policy`,
            type: "TAG_POLICY",
          }
        );

        new OrganizationsPolicyAttachment(
          this,
          `env_tag_policy_attach_${acct.name}`,
          {
            policyId: acctEnvTagPolicy.id,
            targetId: account.id,
          }
        );

        return { account, name: acct.name };
      });
    });

    new TerraformOutput(this, "org-last-updated", {
      description:
        "The timestamp of when the module was last applied. Useful for forcing applies on upgrade",
      value: new Date().toISOString(),
    });
  }
}
