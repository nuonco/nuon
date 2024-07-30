import { DataAwsCallerIdentity } from "@cdktf/provider-aws/lib/data-aws-caller-identity";
import { DataAwsRoute53Zone } from "@cdktf/provider-aws/lib/data-aws-route53-zone";
import { IamAccountAlias } from "@cdktf/provider-aws/lib/iam-account-alias";
import { IamOpenidConnectProvider } from "@cdktf/provider-aws/lib/iam-openid-connect-provider";
import { IamRole } from "@cdktf/provider-aws/lib/iam-role";
import { IamServiceLinkedRole } from "@cdktf/provider-aws/lib/iam-service-linked-role";
import {
  AwsProvider,
  AwsProviderDefaultTags,
} from "@cdktf/provider-aws/lib/provider";
import { Route53Record } from "@cdktf/provider-aws/lib/route53-record";
import { Route53Zone } from "@cdktf/provider-aws/lib/route53-zone";
import { TerraformOutput, TerraformStack } from "cdktf";
import { Construct } from "constructs";
import { allowedRegions, defaultRegion, rootDomain } from "./defaults";
import { TAWSAccount } from "./org";

interface IAccountsConfig {
  accounts: TAWSAccount[];
  defaultTags: AwsProviderDefaultTags[];
}

// Accounts generates the terraform that configures all accounts.
// This is just account level configuration. Regional configuration will
// be added in a different class at a later date.
export class Accounts extends TerraformStack {
  constructor(scope: Construct, name: string, config: IAccountsConfig) {
    super(scope, name);

    const mgmtProvider = new AwsProvider(this, "accounts", {
      defaultTags: config.defaultTags,
      region: defaultRegion,
    });

    const mgmtAcctId = new DataAwsCallerIdentity(this, "mgmt-acct-identity", {
      provider: mgmtProvider,
    }).id;

    const nuonCoZone = new DataAwsRoute53Zone(this, `accounts-nuon-zone`, {
      name: `${rootDomain}.`,
      provider: mgmtProvider,
    });

    config.accounts.forEach((acct) => {
      const defaultRegionalProvider = new AwsProvider(
        this,
        `accounts-${acct.name}-provider`,
        {
          alias: acct.name,
          assumeRole: [
            {
              roleArn: `arn:aws:iam::${acct.account.id}:role/OrganizationAccountAccessRole`,
            },
          ],
          defaultTags: config.defaultTags,
          region: defaultRegion,
        }
      );

      allowedRegions.reduce<Record<string, AwsProvider>>((accum, region) => {
        return {
          ...accum,
          ...{
            [region]: new AwsProvider(
              this,
              `accounts-${acct.name}-provider-${region}`,
              {
                alias: `${acct.name}-${region}`,
                assumeRole: [
                  {
                    roleArn: `arn:aws:iam::${acct.account.id}:role/OrganizationAccountAccessRole`,
                  },
                ],
                defaultTags: config.defaultTags,
                region,
              }
            ),
          },
        };
      }, {});

      new IamRole(this, `${acct.name}-terraform-role`, {
        assumeRolePolicy: JSON.stringify({
          Statement: {
            Action: "sts:AssumeRole",
            Effect: "Allow",
            Principal: {
              AWS: `arn:aws:iam::${mgmtAcctId}:root`,
            },
          },
          Version: "2012-10-17",
        }),
        description: "Terraform role to assume",
        managedPolicyArns: ["arn:aws:iam::aws:policy/AdministratorAccess"],
        name: "terraform",
        path: "/",
        provider: defaultRegionalProvider,
      });

      // Github OIDC provider. This will allow actions to authenticate to our accounts.
      // Currently, there are no roles for it to assume.
      const ghoidc = new IamOpenidConnectProvider(
        this,
        `${acct.name}-github-oidc-provider`,
        {
          clientIdList: ["sts.amazonaws.com"],
          provider: defaultRegionalProvider,
          thumbprintList: [
            "6938fd4d98bab03faadb97b34396831e3780aea1",
            "1c58a3a8518e8759bf075b76b750d4f2df264fcd",
          ],
          url: "https://token.actions.githubusercontent.com",
        }
      );

      new TerraformOutput(this, `${acct.name}-gh-oidc-provider`, {
        description: `The Github OIDC provider for the ${acct.name} account`,
        value: ghoidc,
      });

      new IamServiceLinkedRole(this, `${acct.name}-spot-service-linked-role`, {
        awsServiceName: "spot.amazonaws.com",
        provider: defaultRegionalProvider,
      });

      new IamServiceLinkedRole(
        this,
        `${acct.name}-spot-fleet-service-linked-role`,
        {
          awsServiceName: "spotfleet.amazonaws.com",
          provider: defaultRegionalProvider,
        }
      );

      new IamServiceLinkedRole(
        this,
        `${acct.name}-opensearch-service-linked-role`,
        {
          awsServiceName: "opensearchservice.amazonaws.com",
          provider: defaultRegionalProvider,
        }
      );

      // an account alias is required for e.g. aws-nuke to work
      new IamAccountAlias(this, `${acct.name}-iam-alias`, {
        accountAlias: `nuon-${acct.name}`,
        provider: defaultRegionalProvider,
      });

      const subdomain = `${acct.name}.${rootDomain}`;

      // we want to delegate e.g. staging.nuon.co to the respoective account
      // so create hosted zone in subaccount
      const subZone = new Route53Zone(this, `${acct.name}-hosted-zone`, {
        name: subdomain,
        provider: defaultRegionalProvider,
      });

      // and create the NS record to delegate
      new Route53Record(this, `${acct.name}-ns-delegation`, {
        name: subdomain,
        provider: mgmtProvider,
        records: subZone.nameServers,
        ttl: 300,
        type: "NS",
        zoneId: nuonCoZone.zoneId,
      });
    });

    new TerraformOutput(this, "accounts-last-updated", {
      description:
        "The timestamp of when the module was last applied. Useful for forcing applies on upgrade",
      value: new Date().toISOString(),
    });
  }
}
