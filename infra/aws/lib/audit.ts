import { Cloudtrail } from "@cdktf/provider-aws/lib/cloudtrail";
import { DataAwsCallerIdentity } from "@cdktf/provider-aws/lib/data-aws-caller-identity";
import { DataAwsIamPolicyDocument } from "@cdktf/provider-aws/lib/data-aws-iam-policy-document";
import { OrganizationsOrganization } from "@cdktf/provider-aws/lib/organizations-organization";
import {
  AwsProvider,
  AwsProviderDefaultTags,
} from "@cdktf/provider-aws/lib/provider";
import { S3Bucket } from "@cdktf/provider-aws/lib/s3-bucket";
import { S3BucketLifecycleConfiguration } from "@cdktf/provider-aws/lib/s3-bucket-lifecycle-configuration";
import { S3BucketPolicy } from "@cdktf/provider-aws/lib/s3-bucket-policy";
import { S3BucketServerSideEncryptionConfigurationA } from "@cdktf/provider-aws/lib/s3-bucket-server-side-encryption-configuration";
import { TerraformOutput, TerraformStack } from "cdktf";
import { Construct } from "constructs";
import { defaultRegion } from "./defaults";
import { TAWSAccount } from "./org";

interface IAuditConfig {
  account: TAWSAccount;
  defaultTags: AwsProviderDefaultTags[];
  org: OrganizationsOrganization;
}

// Audit generates the terraform resources for our audit config
export class Audit extends TerraformStack {
  constructor(scope: Construct, name: string, config: IAuditConfig) {
    super(scope, name);

    const auditTrailName = "nuon-org-trail";

    const auditProvider = new AwsProvider(this, "audit", {
      alias: "audit",
      assumeRole: [
        {
          roleArn: `arn:aws:iam::${config.account.account.id}:role/OrganizationAccountAccessRole`,
        },
      ],
      defaultTags: config.defaultTags,
      region: defaultRegion,
    });

    const mgmtProvider = new AwsProvider(this, "mgmt", {
      defaultTags: config.defaultTags,
      region: defaultRegion,
    });

    const mgmtAcctId = new DataAwsCallerIdentity(this, "mgmt-acct-identity", {
      provider: mgmtProvider,
    }).id;

    // this is the bucket that will hold all of the audit logs
    const awsS3BucketAudit = new S3Bucket(this, "audit-bucket", {
      bucket: "nuon-audit-bucket",
      provider: auditProvider,
    });

    // only store logs for 1 year. move to IA after 30 days and glacier after 60
    new S3BucketLifecycleConfiguration(this, "audit-bucket-lifecycle", {
      bucket: awsS3BucketAudit.bucket,
      provider: auditProvider,
      rule: [
        {
          expiration: { days: 365 * 1 },
          id: "default-transition-rule",
          status: "Enabled",
          transition: [
            { days: 30, storageClass: "STANDARD_IA" },
            { days: 60, storageClass: "GLACIER" },
          ],
        },
      ],
    });

    // use bucket key for encryption. this is cheaper and reduces risk of KMS rate limiting
    new S3BucketServerSideEncryptionConfigurationA(
      this,
      "audit-bucket-encryption",
      {
        bucket: awsS3BucketAudit.bucket,
        provider: auditProvider,
        rule: [
          {
            applyServerSideEncryptionByDefault: { sseAlgorithm: "aws:kms" },
            bucketKeyEnabled: true,
          },
        ],
      }
    );

    // bucket policy to allow cloudtrail in all accounts
    const trailDerivedARN = `arn:aws:cloudtrail:${defaultRegion}:${mgmtAcctId}:trail/${auditTrailName}`;
    const dataAwsIamPolicyDocumentAuditBucket = new DataAwsIamPolicyDocument(
      this,
      "audit_bucket",
      {
        statement: [
          {
            actions: ["s3:GetBucketAcl"],
            principals: [
              {
                identifiers: ["cloudtrail.amazonaws.com"],
                type: "Service",
              },
            ],
            resources: [awsS3BucketAudit.arn],
            sid: "AWSCloudTrailAclCheck20150319",
          },
          {
            actions: ["s3:PutObject"],
            condition: [
              {
                test: "StringEquals",
                values: [trailDerivedARN],
                variable: "aws:SourceArn",
              },
              {
                test: "StringEquals",
                values: ["bucket-owner-full-control"],
                variable: "s3:x-amz-acl",
              },
            ],
            principals: [
              {
                identifiers: ["cloudtrail.amazonaws.com"],
                type: "Service",
              },
            ],
            resources: [`${awsS3BucketAudit.arn}/AWSLogs/${mgmtAcctId}/*`],
            sid: "AWSCloudTrailWrite20150319Mgmt",
          },
          {
            actions: ["s3:PutObject"],
            condition: [
              {
                test: "StringEquals",
                values: [trailDerivedARN],
                variable: "aws:SourceArn",
              },
              {
                test: "StringEquals",
                values: ["bucket-owner-full-control"],
                variable: "s3:x-amz-acl",
              },
            ],
            principals: [
              {
                identifiers: ["cloudtrail.amazonaws.com"],
                type: "Service",
              },
            ],
            resources: [`${awsS3BucketAudit.arn}/AWSLogs/${config.org.id}/*`],
            sid: "AWSCloudTrailWrite20150319Org",
          },
        ],
      }
    );

    const awsS3BucketPolicyAudit = new S3BucketPolicy(
      this,
      "audit-bucket-policy-attachment",
      {
        bucket: awsS3BucketAudit.id,
        policy: dataAwsIamPolicyDocumentAuditBucket.json,
        provider: auditProvider,
      }
    );

    new Cloudtrail(this, "audit-cloudtrail", {
      dependsOn: [awsS3BucketPolicyAudit],
      includeGlobalServiceEvents: true,
      isMultiRegionTrail: true,
      isOrganizationTrail: true,
      name: auditTrailName,
      s3BucketName: awsS3BucketAudit.bucket,
    });

    new TerraformOutput(this, "audit-last-updated", {
      description:
        "The timestamp of when the module was last applied. Useful for forcing applies on upgrade",
      value: new Date().toISOString(),
    });
  }
}
