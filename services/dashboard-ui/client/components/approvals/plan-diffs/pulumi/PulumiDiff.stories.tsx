export default {
  title: 'Approvals/PlanDiffs/PulumiDiff',
}

import { PulumiDiff } from './PulumiDiff'

export const S3BucketCreate = () => (
  <PulumiDiff
    plan={{
      stdout: '',
      stderr: '',
      change_summary: { create: 3 },
      resource_changes: [
        {
          urn: 'urn:pulumi:prod::data-platform::aws:s3/bucket:Bucket::artifacts-bucket',
          type: 'aws:s3/bucket:Bucket',
          name: 'artifacts-bucket',
          action: 'create',
          new_inputs: {
            bucket: 'acme-artifacts-prod',
            forceDestroy: false,
            tags: { Environment: 'prod', Team: 'platform' },
          },
        },
        {
          urn: 'urn:pulumi:prod::data-platform::aws:s3/bucketPolicy:BucketPolicy::artifacts-bucket-policy',
          type: 'aws:s3/bucketPolicy:BucketPolicy',
          name: 'artifacts-bucket-policy',
          action: 'create',
          new_inputs: {
            bucket: 'acme-artifacts-prod',
            policy: JSON.stringify({
              Version: '2012-10-17',
              Statement: [
                {
                  Effect: 'Allow',
                  Principal: { AWS: 'arn:aws:iam::123456789012:root' },
                  Action: ['s3:GetObject', 's3:PutObject'],
                  Resource: 'arn:aws:s3:::acme-artifacts-prod/*',
                },
              ],
            }),
          },
        },
        {
          urn: 'urn:pulumi:prod::data-platform::aws:s3/bucketNotification:BucketNotification::artifacts-bucket-notification',
          type: 'aws:s3/bucketNotification:BucketNotification',
          name: 'artifacts-bucket-notification',
          action: 'create',
          new_inputs: {
            bucket: 'acme-artifacts-prod',
            lambdaFunctions: [
              {
                lambdaFunctionArn:
                  'arn:aws:lambda:us-east-1:123456789012:function:process-upload',
                events: ['s3:ObjectCreated:*'],
                filterPrefix: 'uploads/',
              },
            ],
          },
        },
      ],
    }}
  />
)

export const ECSServiceUpdate = () => (
  <PulumiDiff
    plan={{
      stdout: '',
      stderr: '',
      change_summary: { update: 3 },
      resource_changes: [
        {
          urn: 'urn:pulumi:prod::api-service::aws:ecs/taskDefinition:TaskDefinition::api-task',
          type: 'aws:ecs/taskDefinition:TaskDefinition',
          name: 'api-task',
          action: 'update',
          diffs: ['containerDefinitions'],
          detailed_diff: {
            'containerDefinitions[0].image': {
              kind: 'update',
              inputDiff: true,
            },
            'containerDefinitions[0].environment[2].value': {
              kind: 'update',
              inputDiff: true,
            },
          },
          old_inputs: {
            containerDefinitions: [
              {
                name: 'api',
                image: '123456789012.dkr.ecr.us-east-1.amazonaws.com/api:v1.8.3',
              },
            ],
          },
          new_inputs: {
            containerDefinitions: [
              {
                name: 'api',
                image: '123456789012.dkr.ecr.us-east-1.amazonaws.com/api:v1.9.0',
              },
            ],
          },
        },
        {
          urn: 'urn:pulumi:prod::api-service::aws:ecs/service:Service::api-service',
          type: 'aws:ecs/service:Service',
          name: 'api-service',
          action: 'update',
          diffs: ['desiredCount'],
          detailed_diff: {
            desiredCount: { kind: 'update', inputDiff: true },
          },
          old_inputs: { desiredCount: 2 },
          new_inputs: { desiredCount: 4 },
        },
        {
          urn: 'urn:pulumi:prod::api-service::aws:cloudwatch/logGroup:LogGroup::api-logs',
          type: 'aws:cloudwatch/logGroup:LogGroup',
          name: 'api-logs',
          action: 'update',
          diffs: ['retentionInDays'],
          detailed_diff: {
            retentionInDays: { kind: 'update', inputDiff: true },
            'tags.updated_by': { kind: 'add', inputDiff: true },
          },
          old_inputs: { retentionInDays: 30 },
          new_inputs: { retentionInDays: 90 },
        },
      ],
    }}
  />
)

export const DatabaseReplace = () => (
  <PulumiDiff
    plan={{
      stdout: '',
      stderr: '',
      change_summary: { 'create-replacement': 1, 'delete-replaced': 1, replace: 1 },
      resource_changes: [
        {
          urn: 'urn:pulumi:prod::database::aws:rds/instance:Instance::primary-db',
          type: 'aws:rds/instance:Instance',
          name: 'primary-db',
          action: 'replace',
          diffs: ['engineVersion'],
          detailed_diff: {
            engineVersion: { kind: 'update', inputDiff: true },
          },
          old_inputs: {
            engineVersion: '14.9',
            instanceClass: 'db.r6g.xlarge',
            allocatedStorage: 100,
          },
          new_inputs: {
            engineVersion: '15.4',
            instanceClass: 'db.r6g.xlarge',
            allocatedStorage: 100,
          },
        },
        {
          urn: 'urn:pulumi:prod::database::aws:rds/instance:Instance::primary-db-replacement',
          type: 'aws:rds/instance:Instance',
          name: 'primary-db-replacement',
          action: 'create-replacement',
          new_inputs: {
            engineVersion: '15.4',
            instanceClass: 'db.r6g.xlarge',
            allocatedStorage: 100,
          },
        },
        {
          urn: 'urn:pulumi:prod::database::aws:rds/instance:Instance::primary-db-old',
          type: 'aws:rds/instance:Instance',
          name: 'primary-db-old',
          action: 'delete-replaced',
          old_inputs: {
            engineVersion: '14.9',
            instanceClass: 'db.r6g.xlarge',
          },
        },
      ],
    }}
  />
)

export const MixedInfraChanges = () => (
  <PulumiDiff
    plan={{
      stdout: '',
      stderr: '',
      change_summary: { same: 1, update: 1, create: 1, delete: 1, read: 1 },
      resource_changes: [
        {
          urn: 'urn:pulumi:prod::networking::aws:ec2/vpc:Vpc::main-vpc',
          type: 'aws:ec2/vpc:Vpc',
          name: 'main-vpc',
          action: 'same',
        },
        {
          urn: 'urn:pulumi:prod::networking::aws:ec2/securityGroup:SecurityGroup::api-sg',
          type: 'aws:ec2/securityGroup:SecurityGroup',
          name: 'api-sg',
          action: 'update',
          diffs: ['ingress'],
          detailed_diff: {
            'ingress[2]': { kind: 'add', inputDiff: true },
            'ingress[0].cidrBlocks[1]': { kind: 'add', inputDiff: true },
          },
          new_inputs: {
            ingress: [
              { fromPort: 443, toPort: 443, protocol: 'tcp', cidrBlocks: ['10.0.0.0/8', '172.16.0.0/12'] },
              { fromPort: 80, toPort: 80, protocol: 'tcp', cidrBlocks: ['10.0.0.0/8'] },
              { fromPort: 8080, toPort: 8080, protocol: 'tcp', cidrBlocks: ['10.0.0.0/8'] },
            ],
          },
        },
        {
          urn: 'urn:pulumi:prod::compute::aws:lambda/function:Function::event-processor',
          type: 'aws:lambda/function:Function',
          name: 'event-processor',
          action: 'create',
          new_inputs: {
            runtime: 'nodejs20.x',
            handler: 'index.handler',
            memorySize: 256,
            timeout: 30,
          },
        },
        {
          urn: 'urn:pulumi:prod::monitoring::aws:cloudwatch/metricAlarm:MetricAlarm::legacy-cpu-alarm',
          type: 'aws:cloudwatch/metricAlarm:MetricAlarm',
          name: 'legacy-cpu-alarm',
          action: 'delete',
          old_inputs: {
            alarmName: 'legacy-cpu-high',
            metricName: 'CPUUtilization',
            threshold: 80,
          },
        },
        {
          urn: 'urn:pulumi:prod::networking::aws:ec2/getSubnets:getSubnets::private-subnets',
          type: 'aws:ec2/getSubnets:getSubnets',
          name: 'private-subnets',
          action: 'read',
        },
      ],
    }}
  />
)

export const WithDiagnostics = () => (
  <PulumiDiff
    plan={{
      stdout: '',
      stderr: '',
      change_summary: { update: 1, create: 1 },
      resource_changes: [
        {
          urn: 'urn:pulumi:prod::api::aws:apigateway/restApi:RestApi::main-api',
          type: 'aws:apigateway/restApi:RestApi',
          name: 'main-api',
          action: 'update',
          diffs: ['description'],
          detailed_diff: {
            description: { kind: 'update', inputDiff: true },
          },
        },
        {
          urn: 'urn:pulumi:prod::api::aws:apigateway/stage:Stage::v2-stage',
          type: 'aws:apigateway/stage:Stage',
          name: 'v2-stage',
          action: 'create',
          new_inputs: {
            stageName: 'v2',
            description: 'V2 API stage',
          },
        },
      ],
      diagnostics: [
        'warning: aws:apigateway/restApi:RestApi (main-api): the "minimumCompressionSize" property is deprecated and will be removed in a future release. Use "minimum_compression_size" instead.',
        'warning: policy-violation: resource aws:apigateway/stage:Stage (v2-stage) is missing required tag "CostCenter". Add the tag to comply with organization policy.',
      ],
    }}
  />
)
