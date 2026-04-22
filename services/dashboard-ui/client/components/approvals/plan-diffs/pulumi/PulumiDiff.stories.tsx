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
                image:
                  '123456789012.dkr.ecr.us-east-1.amazonaws.com/api:v1.8.3',
              },
            ],
          },
          new_inputs: {
            containerDefinitions: [
              {
                name: 'api',
                image:
                  '123456789012.dkr.ecr.us-east-1.amazonaws.com/api:v1.9.0',
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
      change_summary: {
        'create-replacement': 1,
        'delete-replaced': 1,
        replace: 1,
      },
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
              {
                fromPort: 443,
                toPort: 443,
                protocol: 'tcp',
                cidrBlocks: ['10.0.0.0/8', '172.16.0.0/12'],
              },
              {
                fromPort: 80,
                toPort: 80,
                protocol: 'tcp',
                cidrBlocks: ['10.0.0.0/8'],
              },
              {
                fromPort: 8080,
                toPort: 8080,
                protocol: 'tcp',
                cidrBlocks: ['10.0.0.0/8'],
              },
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

export const LargeEKSClusterDeploy = () => (
  <PulumiDiff
    plan={{
      stdout: '',
      stderr: '',
      change_summary: { create: 18, update: 4, same: 6, read: 3 },
      resource_changes: [
        {
          urn: 'urn:pulumi:prod::k8s-platform::eks:index:Cluster::primary',
          type: 'eks:index:Cluster',
          name: 'primary',
          action: 'same',
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:ec2/launchTemplate:LaunchTemplate::workers-general',
          type: 'aws:ec2/launchTemplate:LaunchTemplate',
          name: 'workers-general',
          action: 'update',
          diffs: ['imageId', 'userData', 'blockDeviceMappings'],
          detailed_diff: {
            imageId: { kind: 'update', inputDiff: true },
            userData: { kind: 'update', inputDiff: true },
            'blockDeviceMappings[0].ebs.volumeSize': {
              kind: 'update',
              inputDiff: true,
            },
            'tagSpecifications[0].tags.ami-version': {
              kind: 'update',
              inputDiff: true,
            },
          },
          old_inputs: {
            imageId: 'ami-0a1b2c3d4e5f60001',
            userData: 'base64-encoded-old-userdata',
            blockDeviceMappings: [
              {
                deviceName: '/dev/xvda',
                ebs: {
                  volumeSize: 100,
                  volumeType: 'gp3',
                  iops: 3000,
                  throughput: 125,
                },
              },
            ],
            instanceType: 'c6i.2xlarge',
            tagSpecifications: [
              {
                resourceType: 'instance',
                tags: { 'ami-version': '1.28-2024.01.15', cluster: 'primary' },
              },
            ],
          },
          new_inputs: {
            imageId: 'ami-0f9a8b7c6d5e40002',
            userData: 'base64-encoded-new-userdata',
            blockDeviceMappings: [
              {
                deviceName: '/dev/xvda',
                ebs: {
                  volumeSize: 200,
                  volumeType: 'gp3',
                  iops: 3000,
                  throughput: 125,
                },
              },
            ],
            instanceType: 'c6i.2xlarge',
            tagSpecifications: [
              {
                resourceType: 'instance',
                tags: { 'ami-version': '1.29-2024.06.20', cluster: 'primary' },
              },
            ],
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:eks/nodeGroup:NodeGroup::workers-general',
          type: 'aws:eks/nodeGroup:NodeGroup',
          name: 'workers-general',
          action: 'update',
          diffs: ['scalingConfig', 'version'],
          detailed_diff: {
            'scalingConfig.desiredSize': { kind: 'update', inputDiff: true },
            'scalingConfig.maxSize': { kind: 'update', inputDiff: true },
            version: { kind: 'update', inputDiff: true },
          },
          old_inputs: {
            clusterName: 'primary',
            nodeGroupName: 'workers-general',
            scalingConfig: { desiredSize: 6, minSize: 3, maxSize: 12 },
            version: '1.28',
            instanceTypes: ['c6i.2xlarge'],
          },
          new_inputs: {
            clusterName: 'primary',
            nodeGroupName: 'workers-general',
            scalingConfig: { desiredSize: 10, minSize: 3, maxSize: 20 },
            version: '1.29',
            instanceTypes: ['c6i.2xlarge'],
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:ec2/launchTemplate:LaunchTemplate::workers-gpu',
          type: 'aws:ec2/launchTemplate:LaunchTemplate',
          name: 'workers-gpu',
          action: 'update',
          diffs: ['imageId'],
          detailed_diff: {
            imageId: { kind: 'update', inputDiff: true },
          },
          old_inputs: {
            imageId: 'ami-gpu-old-0001',
            instanceType: 'g5.2xlarge',
          },
          new_inputs: {
            imageId: 'ami-gpu-new-0002',
            instanceType: 'g5.2xlarge',
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:eks/nodeGroup:NodeGroup::workers-gpu',
          type: 'aws:eks/nodeGroup:NodeGroup',
          name: 'workers-gpu',
          action: 'update',
          diffs: ['scalingConfig'],
          detailed_diff: {
            'scalingConfig.desiredSize': { kind: 'update', inputDiff: true },
            'scalingConfig.maxSize': { kind: 'update', inputDiff: true },
          },
          old_inputs: {
            scalingConfig: { desiredSize: 2, minSize: 0, maxSize: 4 },
          },
          new_inputs: {
            scalingConfig: { desiredSize: 4, minSize: 0, maxSize: 8 },
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:ec2/getAmi:getAmi::eks-optimized-ami',
          type: 'aws:ec2/getAmi:getAmi',
          name: 'eks-optimized-ami',
          action: 'read',
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:ec2/getAmi:getAmi::eks-gpu-ami',
          type: 'aws:ec2/getAmi:getAmi',
          name: 'eks-gpu-ami',
          action: 'read',
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:ec2/getSubnets:getSubnets::private-subnets',
          type: 'aws:ec2/getSubnets:getSubnets',
          name: 'private-subnets',
          action: 'read',
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:ec2/vpc:Vpc::main-vpc',
          type: 'aws:ec2/vpc:Vpc',
          name: 'main-vpc',
          action: 'same',
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:iam/role:Role::node-role',
          type: 'aws:iam/role:Role',
          name: 'node-role',
          action: 'same',
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:iam/role:Role::cluster-role',
          type: 'aws:iam/role:Role',
          name: 'cluster-role',
          action: 'same',
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::kubernetes:helm.sh/v4:Chart::karpenter',
          type: 'kubernetes:helm.sh/v4:Chart',
          name: 'karpenter',
          action: 'create',
          new_inputs: {
            chart: 'oci://public.ecr.aws/karpenter/karpenter',
            version: 'v0.35.0',
            namespace: 'karpenter',
            values: {
              settings: {
                clusterName: 'primary',
                clusterEndpoint:
                  'https://ABCDEF1234.gr7.us-east-1.eks.amazonaws.com',
                interruptionQueue: 'primary-karpenter-interruption',
              },
              replicas: 2,
              controller: {
                resources: {
                  requests: { cpu: '1', memory: '1Gi' },
                  limits: { cpu: '2', memory: '2Gi' },
                },
              },
            },
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::kubernetes:core/v1:Namespace::karpenter-ns',
          type: 'kubernetes:core/v1:Namespace',
          name: 'karpenter-ns',
          action: 'create',
          new_inputs: {
            metadata: {
              name: 'karpenter',
              labels: { 'app.kubernetes.io/managed-by': 'pulumi' },
            },
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:iam/role:Role::karpenter-controller',
          type: 'aws:iam/role:Role',
          name: 'karpenter-controller',
          action: 'create',
          new_inputs: {
            name: 'primary-karpenter-controller',
            assumeRolePolicy: JSON.stringify({
              Version: '2012-10-17',
              Statement: [
                {
                  Effect: 'Allow',
                  Principal: {
                    Federated:
                      'arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/ABCDEF1234',
                  },
                  Action: 'sts:AssumeRoleWithWebIdentity',
                  Condition: {
                    StringEquals: {
                      'oidc.eks.us-east-1.amazonaws.com/id/ABCDEF1234:sub':
                        'system:serviceaccount:karpenter:karpenter',
                    },
                  },
                },
              ],
            }),
            tags: { Environment: 'prod', Service: 'karpenter' },
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:iam/rolePolicyAttachment:RolePolicyAttachment::karpenter-controller-policy',
          type: 'aws:iam/rolePolicyAttachment:RolePolicyAttachment',
          name: 'karpenter-controller-policy',
          action: 'create',
          new_inputs: {
            role: 'primary-karpenter-controller',
            policyArn:
              'arn:aws:iam::123456789012:policy/KarpenterControllerPolicy',
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:sqs/queue:Queue::karpenter-interruption',
          type: 'aws:sqs/queue:Queue',
          name: 'karpenter-interruption',
          action: 'create',
          new_inputs: {
            name: 'primary-karpenter-interruption',
            messageRetentionSeconds: 300,
            sqsManagedSseEnabled: true,
            tags: { Environment: 'prod', Service: 'karpenter' },
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:sqs/queuePolicy:QueuePolicy::karpenter-interruption-policy',
          type: 'aws:sqs/queuePolicy:QueuePolicy',
          name: 'karpenter-interruption-policy',
          action: 'create',
          new_inputs: {
            queueUrl:
              'https://sqs.us-east-1.amazonaws.com/123456789012/primary-karpenter-interruption',
            policy: JSON.stringify({
              Version: '2012-10-17',
              Statement: [
                {
                  Effect: 'Allow',
                  Principal: {
                    Service: ['events.amazonaws.com', 'sqs.amazonaws.com'],
                  },
                  Action: 'sqs:SendMessage',
                  Resource:
                    'arn:aws:sqs:us-east-1:123456789012:primary-karpenter-interruption',
                },
              ],
            }),
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:cloudwatch/eventRule:EventRule::spot-interruption',
          type: 'aws:cloudwatch/eventRule:EventRule',
          name: 'spot-interruption',
          action: 'create',
          new_inputs: {
            name: 'primary-karpenter-spot-interruption',
            eventPattern: JSON.stringify({
              source: ['aws.ec2'],
              'detail-type': ['EC2 Spot Instance Interruption Warning'],
            }),
            tags: { Environment: 'prod' },
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:cloudwatch/eventTarget:EventTarget::spot-interruption-target',
          type: 'aws:cloudwatch/eventTarget:EventTarget',
          name: 'spot-interruption-target',
          action: 'create',
          new_inputs: {
            rule: 'primary-karpenter-spot-interruption',
            arn: 'arn:aws:sqs:us-east-1:123456789012:primary-karpenter-interruption',
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:cloudwatch/eventRule:EventRule::instance-rebalance',
          type: 'aws:cloudwatch/eventRule:EventRule',
          name: 'instance-rebalance',
          action: 'create',
          new_inputs: {
            name: 'primary-karpenter-instance-rebalance',
            eventPattern: JSON.stringify({
              source: ['aws.ec2'],
              'detail-type': ['EC2 Instance Rebalance Recommendation'],
            }),
            tags: { Environment: 'prod' },
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:cloudwatch/eventTarget:EventTarget::instance-rebalance-target',
          type: 'aws:cloudwatch/eventTarget:EventTarget',
          name: 'instance-rebalance-target',
          action: 'create',
          new_inputs: {
            rule: 'primary-karpenter-instance-rebalance',
            arn: 'arn:aws:sqs:us-east-1:123456789012:primary-karpenter-interruption',
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::kubernetes:core/v1:Namespace::monitoring-ns',
          type: 'kubernetes:core/v1:Namespace',
          name: 'monitoring-ns',
          action: 'same',
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::kubernetes:helm.sh/v4:Chart::metrics-server',
          type: 'kubernetes:helm.sh/v4:Chart',
          name: 'metrics-server',
          action: 'create',
          new_inputs: {
            chart: 'metrics-server',
            version: '3.12.0',
            repositoryOpts: {
              repo: 'https://kubernetes-sigs.github.io/metrics-server/',
            },
            namespace: 'kube-system',
            values: {
              replicas: 2,
              resources: { requests: { cpu: '100m', memory: '200Mi' } },
            },
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::kubernetes:helm.sh/v4:Chart::cluster-autoscaler',
          type: 'kubernetes:helm.sh/v4:Chart',
          name: 'cluster-autoscaler',
          action: 'create',
          new_inputs: {
            chart: 'cluster-autoscaler',
            version: '9.35.0',
            repositoryOpts: { repo: 'https://kubernetes.github.io/autoscaler' },
            namespace: 'kube-system',
            values: {
              autoDiscovery: { clusterName: 'primary' },
              awsRegion: 'us-east-1',
              extraArgs: {
                'balance-similar-node-groups': true,
                'skip-nodes-with-system-pods': false,
              },
            },
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::kubernetes:helm.sh/v4:Chart::aws-load-balancer-controller',
          type: 'kubernetes:helm.sh/v4:Chart',
          name: 'aws-load-balancer-controller',
          action: 'create',
          new_inputs: {
            chart: 'aws-load-balancer-controller',
            version: '1.7.1',
            repositoryOpts: { repo: 'https://aws.github.io/eks-charts' },
            namespace: 'kube-system',
            values: {
              clusterName: 'primary',
              serviceAccount: {
                create: true,
                name: 'aws-load-balancer-controller',
                annotations: {
                  'eks.amazonaws.com/role-arn':
                    'arn:aws:iam::123456789012:role/primary-alb-controller',
                },
              },
              vpcId: 'vpc-0abc123def456789',
              region: 'us-east-1',
            },
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::aws:iam/role:Role::alb-controller',
          type: 'aws:iam/role:Role',
          name: 'alb-controller',
          action: 'create',
          new_inputs: {
            name: 'primary-alb-controller',
            assumeRolePolicy: JSON.stringify({
              Version: '2012-10-17',
              Statement: [
                {
                  Effect: 'Allow',
                  Principal: {
                    Federated:
                      'arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/ABCDEF1234',
                  },
                  Action: 'sts:AssumeRoleWithWebIdentity',
                },
              ],
            }),
            tags: { Environment: 'prod', Service: 'alb-controller' },
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::kubernetes:helm.sh/v4:Chart::external-dns',
          type: 'kubernetes:helm.sh/v4:Chart',
          name: 'external-dns',
          action: 'create',
          new_inputs: {
            chart: 'external-dns',
            version: '1.14.3',
            repositoryOpts: {
              repo: 'https://kubernetes-sigs.github.io/external-dns/',
            },
            namespace: 'kube-system',
            values: {
              provider: 'aws',
              domainFilters: ['prod.acme.io'],
              policy: 'sync',
              registry: 'txt',
              txtOwnerId: 'primary-cluster',
              serviceAccount: {
                annotations: {
                  'eks.amazonaws.com/role-arn':
                    'arn:aws:iam::123456789012:role/primary-external-dns',
                },
              },
            },
          },
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::kubernetes:core/v1:Namespace::cert-manager-ns',
          type: 'kubernetes:core/v1:Namespace',
          name: 'cert-manager-ns',
          action: 'same',
        },
        {
          urn: 'urn:pulumi:prod::k8s-platform::kubernetes:helm.sh/v4:Chart::cert-manager',
          type: 'kubernetes:helm.sh/v4:Chart',
          name: 'cert-manager',
          action: 'create',
          new_inputs: {
            chart: 'cert-manager',
            version: 'v1.14.4',
            repositoryOpts: { repo: 'https://charts.jetstack.io' },
            namespace: 'cert-manager',
            values: {
              installCRDs: true,
              replicaCount: 2,
              webhook: { replicaCount: 2 },
              cainjector: { replicaCount: 2 },
            },
          },
        },
      ],
    }}
  />
)

export const KubernetesAppMigration = () => (
  <PulumiDiff
    plan={{
      stdout: '',
      stderr: '',
      change_summary: {
        create: 12,
        update: 5,
        delete: 3,
        replace: 2,
        'create-replacement': 2,
        'delete-replaced': 2,
        same: 8,
        read: 2,
      },
      resource_changes: [
        {
          urn: 'urn:pulumi:prod::order-service::kubernetes:core/v1:Namespace::order-service-ns',
          type: 'kubernetes:core/v1:Namespace',
          name: 'order-service-ns',
          action: 'same',
        },
        {
          urn: 'urn:pulumi:prod::order-service::kubernetes:apps/v1:Deployment::order-api',
          type: 'kubernetes:apps/v1:Deployment',
          name: 'order-api',
          action: 'update',
          diffs: ['spec'],
          detailed_diff: {
            'spec.template.spec.containers[0].image': {
              kind: 'update',
              inputDiff: true,
            },
            'spec.template.spec.containers[0].resources.requests.memory': {
              kind: 'update',
              inputDiff: true,
            },
            'spec.template.spec.containers[0].resources.limits.memory': {
              kind: 'update',
              inputDiff: true,
            },
            'spec.template.spec.containers[0].env[3].value': {
              kind: 'update',
              inputDiff: true,
            },
            'spec.template.spec.containers[0].env[4]': {
              kind: 'add',
              inputDiff: true,
            },
            'spec.replicas': { kind: 'update', inputDiff: true },
            'spec.template.metadata.annotations.configHash': {
              kind: 'update',
              inputDiff: true,
            },
          },
          old_inputs: {
            metadata: { name: 'order-api', namespace: 'order-service' },
            spec: {
              replicas: 3,
              template: {
                spec: {
                  containers: [
                    {
                      name: 'order-api',
                      image:
                        '123456789012.dkr.ecr.us-east-1.amazonaws.com/order-api:v2.14.0',
                      resources: {
                        requests: { cpu: '500m', memory: '512Mi' },
                        limits: { cpu: '1000m', memory: '1Gi' },
                      },
                      env: [
                        {
                          name: 'DATABASE_URL',
                          valueFrom: {
                            secretKeyRef: {
                              name: 'order-db-credentials',
                              key: 'url',
                            },
                          },
                        },
                        {
                          name: 'REDIS_URL',
                          value: 'redis://order-cache.internal:6379',
                        },
                        { name: 'LOG_LEVEL', value: 'info' },
                        {
                          name: 'FEATURE_FLAGS_ENDPOINT',
                          value: 'http://flagsmith.internal:8000/api/v1',
                        },
                      ],
                    },
                  ],
                },
              },
            },
          },
          new_inputs: {
            metadata: { name: 'order-api', namespace: 'order-service' },
            spec: {
              replicas: 5,
              template: {
                spec: {
                  containers: [
                    {
                      name: 'order-api',
                      image:
                        '123456789012.dkr.ecr.us-east-1.amazonaws.com/order-api:v3.0.0',
                      resources: {
                        requests: { cpu: '500m', memory: '1Gi' },
                        limits: { cpu: '1000m', memory: '2Gi' },
                      },
                      env: [
                        {
                          name: 'DATABASE_URL',
                          valueFrom: {
                            secretKeyRef: {
                              name: 'order-db-credentials',
                              key: 'url',
                            },
                          },
                        },
                        {
                          name: 'REDIS_URL',
                          value: 'redis://order-cache.internal:6379',
                        },
                        { name: 'LOG_LEVEL', value: 'info' },
                        {
                          name: 'FEATURE_FLAGS_ENDPOINT',
                          value: 'http://flagsmith.internal:8100/api/v2',
                        },
                        {
                          name: 'OTEL_EXPORTER_OTLP_ENDPOINT',
                          value: 'http://otel-collector.monitoring:4317',
                        },
                      ],
                    },
                  ],
                },
              },
            },
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::kubernetes:apps/v1:Deployment::order-worker',
          type: 'kubernetes:apps/v1:Deployment',
          name: 'order-worker',
          action: 'update',
          diffs: ['spec'],
          detailed_diff: {
            'spec.template.spec.containers[0].image': {
              kind: 'update',
              inputDiff: true,
            },
            'spec.template.spec.containers[0].args[0]': {
              kind: 'update',
              inputDiff: true,
            },
            'spec.replicas': { kind: 'update', inputDiff: true },
          },
          old_inputs: {
            spec: {
              replicas: 2,
              template: {
                spec: {
                  containers: [
                    {
                      name: 'order-worker',
                      image:
                        '123456789012.dkr.ecr.us-east-1.amazonaws.com/order-worker:v2.14.0',
                      args: ['--concurrency=10'],
                    },
                  ],
                },
              },
            },
          },
          new_inputs: {
            spec: {
              replicas: 4,
              template: {
                spec: {
                  containers: [
                    {
                      name: 'order-worker',
                      image:
                        '123456789012.dkr.ecr.us-east-1.amazonaws.com/order-worker:v3.0.0',
                      args: ['--concurrency=20'],
                    },
                  ],
                },
              },
            },
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::kubernetes:core/v1:Service::order-api-svc',
          type: 'kubernetes:core/v1:Service',
          name: 'order-api-svc',
          action: 'same',
        },
        {
          urn: 'urn:pulumi:prod::order-service::kubernetes:core/v1:ConfigMap::order-api-config',
          type: 'kubernetes:core/v1:ConfigMap',
          name: 'order-api-config',
          action: 'update',
          diffs: ['data'],
          detailed_diff: {
            'data.MAX_CONNECTIONS': { kind: 'update', inputDiff: true },
            'data.CACHE_TTL': { kind: 'update', inputDiff: true },
            'data.OTEL_SERVICE_NAME': { kind: 'add', inputDiff: true },
            'data.OTEL_TRACES_SAMPLER_ARG': { kind: 'add', inputDiff: true },
          },
          old_inputs: {
            data: {
              MAX_CONNECTIONS: '100',
              CACHE_TTL: '300',
              RATE_LIMIT: '1000',
            },
          },
          new_inputs: {
            data: {
              MAX_CONNECTIONS: '200',
              CACHE_TTL: '600',
              RATE_LIMIT: '1000',
              OTEL_SERVICE_NAME: 'order-api',
              OTEL_TRACES_SAMPLER_ARG: '0.1',
            },
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:rds/instance:Instance::order-db',
          type: 'aws:rds/instance:Instance',
          name: 'order-db',
          action: 'replace',
          diffs: ['instanceClass', 'engineVersion'],
          detailed_diff: {
            instanceClass: { kind: 'update', inputDiff: true },
            engineVersion: { kind: 'update', inputDiff: true },
            performanceInsightsEnabled: { kind: 'add', inputDiff: true },
          },
          old_inputs: {
            identifier: 'order-db-prod',
            instanceClass: 'db.r6g.large',
            engineVersion: '15.4',
            engine: 'postgres',
            allocatedStorage: 200,
            multiAz: true,
          },
          new_inputs: {
            identifier: 'order-db-prod',
            instanceClass: 'db.r6g.xlarge',
            engineVersion: '16.1',
            engine: 'postgres',
            allocatedStorage: 200,
            multiAz: true,
            performanceInsightsEnabled: true,
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:rds/instance:Instance::order-db-replacement',
          type: 'aws:rds/instance:Instance',
          name: 'order-db-replacement',
          action: 'create-replacement',
          new_inputs: {
            identifier: 'order-db-prod',
            instanceClass: 'db.r6g.xlarge',
            engineVersion: '16.1',
            engine: 'postgres',
            allocatedStorage: 200,
            multiAz: true,
            performanceInsightsEnabled: true,
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:rds/instance:Instance::order-db-old',
          type: 'aws:rds/instance:Instance',
          name: 'order-db-old',
          action: 'delete-replaced',
          old_inputs: {
            identifier: 'order-db-prod',
            instanceClass: 'db.r6g.large',
            engineVersion: '15.4',
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:elasticache/replicationGroup:ReplicationGroup::order-cache',
          type: 'aws:elasticache/replicationGroup:ReplicationGroup',
          name: 'order-cache',
          action: 'replace',
          diffs: ['engineVersion', 'nodeType'],
          detailed_diff: {
            engineVersion: { kind: 'update', inputDiff: true },
            nodeType: { kind: 'update', inputDiff: true },
          },
          old_inputs: {
            replicationGroupId: 'order-cache-prod',
            engineVersion: '7.0',
            nodeType: 'cache.r6g.large',
            numCacheClusters: 3,
          },
          new_inputs: {
            replicationGroupId: 'order-cache-prod',
            engineVersion: '7.1',
            nodeType: 'cache.r7g.large',
            numCacheClusters: 3,
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:elasticache/replicationGroup:ReplicationGroup::order-cache-replacement',
          type: 'aws:elasticache/replicationGroup:ReplicationGroup',
          name: 'order-cache-replacement',
          action: 'create-replacement',
          new_inputs: {
            replicationGroupId: 'order-cache-prod',
            engineVersion: '7.1',
            nodeType: 'cache.r7g.large',
            numCacheClusters: 3,
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:elasticache/replicationGroup:ReplicationGroup::order-cache-old',
          type: 'aws:elasticache/replicationGroup:ReplicationGroup',
          name: 'order-cache-old',
          action: 'delete-replaced',
          old_inputs: {
            replicationGroupId: 'order-cache-prod',
            engineVersion: '7.0',
            nodeType: 'cache.r6g.large',
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::kubernetes:autoscaling/v2:HorizontalPodAutoscaler::order-api-hpa',
          type: 'kubernetes:autoscaling/v2:HorizontalPodAutoscaler',
          name: 'order-api-hpa',
          action: 'update',
          diffs: ['spec'],
          detailed_diff: {
            'spec.maxReplicas': { kind: 'update', inputDiff: true },
            'spec.metrics[1]': { kind: 'add', inputDiff: true },
          },
          old_inputs: {
            spec: {
              minReplicas: 3,
              maxReplicas: 10,
              metrics: [
                {
                  type: 'Resource',
                  resource: {
                    name: 'cpu',
                    target: { type: 'Utilization', averageUtilization: 70 },
                  },
                },
              ],
            },
          },
          new_inputs: {
            spec: {
              minReplicas: 5,
              maxReplicas: 20,
              metrics: [
                {
                  type: 'Resource',
                  resource: {
                    name: 'cpu',
                    target: { type: 'Utilization', averageUtilization: 70 },
                  },
                },
                {
                  type: 'Resource',
                  resource: {
                    name: 'memory',
                    target: { type: 'Utilization', averageUtilization: 80 },
                  },
                },
              ],
            },
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::kubernetes:networking.k8s.io/v1:NetworkPolicy::order-api-netpol',
          type: 'kubernetes:networking.k8s.io/v1:NetworkPolicy',
          name: 'order-api-netpol',
          action: 'create',
          new_inputs: {
            metadata: { name: 'order-api-netpol', namespace: 'order-service' },
            spec: {
              podSelector: { matchLabels: { app: 'order-api' } },
              policyTypes: ['Ingress', 'Egress'],
              ingress: [
                {
                  from: [
                    {
                      namespaceSelector: {
                        matchLabels: {
                          'kubernetes.io/metadata.name': 'ingress-nginx',
                        },
                      },
                    },
                  ],
                  ports: [{ port: 8080, protocol: 'TCP' }],
                },
              ],
              egress: [
                {
                  to: [{ namespaceSelector: {} }],
                  ports: [
                    { port: 5432, protocol: 'TCP' },
                    { port: 6379, protocol: 'TCP' },
                  ],
                },
              ],
            },
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::kubernetes:policy/v1:PodDisruptionBudget::order-api-pdb',
          type: 'kubernetes:policy/v1:PodDisruptionBudget',
          name: 'order-api-pdb',
          action: 'create',
          new_inputs: {
            metadata: { name: 'order-api-pdb', namespace: 'order-service' },
            spec: {
              minAvailable: '50%',
              selector: { matchLabels: { app: 'order-api' } },
            },
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::kubernetes:monitoring.coreos.com/v1:ServiceMonitor::order-api-monitor',
          type: 'kubernetes:monitoring.coreos.com/v1:ServiceMonitor',
          name: 'order-api-monitor',
          action: 'create',
          new_inputs: {
            metadata: {
              name: 'order-api-monitor',
              namespace: 'order-service',
              labels: { release: 'prometheus' },
            },
            spec: {
              selector: { matchLabels: { app: 'order-api' } },
              endpoints: [
                { port: 'metrics', interval: '15s', path: '/metrics' },
              ],
            },
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::kubernetes:monitoring.coreos.com/v1:PrometheusRule::order-api-alerts',
          type: 'kubernetes:monitoring.coreos.com/v1:PrometheusRule',
          name: 'order-api-alerts',
          action: 'create',
          new_inputs: {
            metadata: { name: 'order-api-alerts', namespace: 'order-service' },
            spec: {
              groups: [
                {
                  name: 'order-api.rules',
                  rules: [
                    {
                      alert: 'OrderAPIHighErrorRate',
                      expr: 'rate(http_requests_total{service="order-api",code=~"5.."}[5m]) / rate(http_requests_total{service="order-api"}[5m]) > 0.05',
                      for: '5m',
                      labels: { severity: 'critical' },
                      annotations: { summary: 'Order API error rate above 5%' },
                    },
                    {
                      alert: 'OrderAPIHighLatency',
                      expr: 'histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{service="order-api"}[5m])) > 2',
                      for: '10m',
                      labels: { severity: 'warning' },
                      annotations: {
                        summary: 'Order API p99 latency above 2s',
                      },
                    },
                  ],
                },
              ],
            },
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:cloudwatch/metricAlarm:MetricAlarm::order-db-cpu',
          type: 'aws:cloudwatch/metricAlarm:MetricAlarm',
          name: 'order-db-cpu',
          action: 'create',
          new_inputs: {
            alarmName: 'order-db-prod-cpu-high',
            metricName: 'CPUUtilization',
            namespace: 'AWS/RDS',
            statistic: 'Average',
            period: 300,
            threshold: 80,
            comparisonOperator: 'GreaterThanThreshold',
            evaluationPeriods: 3,
            dimensions: { DBInstanceIdentifier: 'order-db-prod' },
            alarmActions: ['arn:aws:sns:us-east-1:123456789012:ops-alerts'],
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:cloudwatch/metricAlarm:MetricAlarm::order-db-connections',
          type: 'aws:cloudwatch/metricAlarm:MetricAlarm',
          name: 'order-db-connections',
          action: 'create',
          new_inputs: {
            alarmName: 'order-db-prod-connections-high',
            metricName: 'DatabaseConnections',
            namespace: 'AWS/RDS',
            statistic: 'Average',
            period: 300,
            threshold: 150,
            comparisonOperator: 'GreaterThanThreshold',
            evaluationPeriods: 2,
            dimensions: { DBInstanceIdentifier: 'order-db-prod' },
            alarmActions: ['arn:aws:sns:us-east-1:123456789012:ops-alerts'],
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:cloudwatch/logGroup:LogGroup::order-api-logs',
          type: 'aws:cloudwatch/logGroup:LogGroup',
          name: 'order-api-logs',
          action: 'same',
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:ec2/securityGroup:SecurityGroup::order-db-sg',
          type: 'aws:ec2/securityGroup:SecurityGroup',
          name: 'order-db-sg',
          action: 'same',
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:ec2/securityGroup:SecurityGroup::order-cache-sg',
          type: 'aws:ec2/securityGroup:SecurityGroup',
          name: 'order-cache-sg',
          action: 'same',
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:iam/role:Role::order-api-task-role',
          type: 'aws:iam/role:Role',
          name: 'order-api-task-role',
          action: 'same',
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:ec2/getSubnets:getSubnets::app-subnets',
          type: 'aws:ec2/getSubnets:getSubnets',
          name: 'app-subnets',
          action: 'read',
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:ec2/getSecurityGroup:getSecurityGroup::vpc-default-sg',
          type: 'aws:ec2/getSecurityGroup:getSecurityGroup',
          name: 'vpc-default-sg',
          action: 'read',
        },
        {
          urn: 'urn:pulumi:prod::order-service::kubernetes:core/v1:Secret::order-legacy-api-key',
          type: 'kubernetes:core/v1:Secret',
          name: 'order-legacy-api-key',
          action: 'delete',
          old_inputs: {
            metadata: {
              name: 'order-legacy-api-key',
              namespace: 'order-service',
            },
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:cloudwatch/metricAlarm:MetricAlarm::order-legacy-alarm',
          type: 'aws:cloudwatch/metricAlarm:MetricAlarm',
          name: 'order-legacy-alarm',
          action: 'delete',
          old_inputs: {
            alarmName: 'order-legacy-health',
            metricName: 'HealthCheckStatus',
            namespace: 'AWS/Route53',
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:route53/record:Record::order-legacy-dns',
          type: 'aws:route53/record:Record',
          name: 'order-legacy-dns',
          action: 'delete',
          old_inputs: {
            name: 'order-legacy.prod.acme.io',
            type: 'CNAME',
            ttl: 300,
            records: ['order-api-legacy.prod.internal'],
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::kubernetes:core/v1:ServiceAccount::order-api-sa',
          type: 'kubernetes:core/v1:ServiceAccount',
          name: 'order-api-sa',
          action: 'create',
          new_inputs: {
            metadata: {
              name: 'order-api',
              namespace: 'order-service',
              annotations: {
                'eks.amazonaws.com/role-arn':
                  'arn:aws:iam::123456789012:role/order-api-irsa',
              },
            },
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:iam/role:Role::order-api-irsa',
          type: 'aws:iam/role:Role',
          name: 'order-api-irsa',
          action: 'create',
          new_inputs: {
            name: 'order-api-irsa',
            assumeRolePolicy: JSON.stringify({
              Version: '2012-10-17',
              Statement: [
                {
                  Effect: 'Allow',
                  Principal: {
                    Federated:
                      'arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-east-1.amazonaws.com/id/ABCDEF1234',
                  },
                  Action: 'sts:AssumeRoleWithWebIdentity',
                  Condition: {
                    StringEquals: {
                      'oidc.eks.us-east-1.amazonaws.com/id/ABCDEF1234:sub':
                        'system:serviceaccount:order-service:order-api',
                    },
                  },
                },
              ],
            }),
            tags: { Environment: 'prod', Service: 'order-api' },
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:iam/rolePolicyAttachment:RolePolicyAttachment::order-api-s3-access',
          type: 'aws:iam/rolePolicyAttachment:RolePolicyAttachment',
          name: 'order-api-s3-access',
          action: 'create',
          new_inputs: {
            role: 'order-api-irsa',
            policyArn: 'arn:aws:iam::123456789012:policy/order-api-s3-access',
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::aws:iam/rolePolicyAttachment:RolePolicyAttachment::order-api-sqs-access',
          type: 'aws:iam/rolePolicyAttachment:RolePolicyAttachment',
          name: 'order-api-sqs-access',
          action: 'create',
          new_inputs: {
            role: 'order-api-irsa',
            policyArn: 'arn:aws:iam::123456789012:policy/order-api-sqs-access',
          },
        },
        {
          urn: 'urn:pulumi:prod::order-service::kubernetes:core/v1:Namespace::order-service-jobs-ns',
          type: 'kubernetes:core/v1:Namespace',
          name: 'order-service-jobs-ns',
          action: 'same',
        },
      ],
      diagnostics: [
        'warning: aws:rds/instance:Instance (order-db): upgrading engineVersion from 15.4 to 16.1 requires replacement. This will cause downtime. Consider using a blue-green deployment strategy.',
        'warning: aws:elasticache/replicationGroup:ReplicationGroup (order-cache): upgrading nodeType from cache.r6g.large to cache.r7g.large requires replacement. Existing connections will be dropped.',
        'warning: policy-violation: resource kubernetes:apps/v1:Deployment (order-api) exceeds recommended memory limit of 1Gi. Current limit: 2Gi. Ensure adequate cluster capacity.',
        'error: preview-only: resource aws:iam/role:Role (order-api-irsa) references OIDC provider that may not exist yet. Verify EKS OIDC provider is configured.',
      ],
    }}
  />
)

export const AzureCosmeticUpdates = () => (
  <PulumiDiff
    plan={{
      stdout: '',
      stderr: '',
      change_summary: { same: 5, update: 3 },
      resource_changes: [
        {
          urn: 'urn:pulumi:prod::infra::azure-native:containerregistry:Registry::acr',
          type: 'azure-native:containerregistry:Registry',
          name: 'acr',
          action: 'update',
          diffs: ['tags'],
          detailed_diff: {
            tags: { kind: 'update', inputDiff: true },
          },
          old_inputs: {
            registryName: 'myregistry',
            sku: { name: 'Premium' },
            location: 'centralindia',
            tags: null,
          },
          new_inputs: {
            registryName: 'myregistry',
            sku: { name: 'Premium' },
            location: 'centralindia',
            tags: {},
          },
        },
        {
          urn: 'urn:pulumi:prod::infra::azure-native:network:Zone::public-zone',
          type: 'azure-native:network:Zone',
          name: 'public-zone',
          action: 'update',
          diffs: ['tags'],
          detailed_diff: {
            tags: { kind: 'update', inputDiff: true },
          },
          old_inputs: {
            zoneName: 'byoc-azure.nuon.co',
            resourceGroupName: 'my-rg',
            tags: null,
          },
          new_inputs: {
            zoneName: 'byoc-azure.nuon.co',
            resourceGroupName: 'my-rg',
            tags: {},
          },
        },
        {
          urn: 'urn:pulumi:prod::infra::azure-native:containerservice:ManagedCluster::aks',
          type: 'azure-native:containerservice:ManagedCluster',
          name: 'aks',
          action: 'update',
          diffs: ['tags', 'agentPoolProfiles', 'identity'],
          detailed_diff: {
            tags: { kind: 'update', inputDiff: true },
            'agentPoolProfiles[0].availabilityZones': {
              kind: 'update',
              inputDiff: true,
            },
            'identity.userAssignedIdentities': {
              kind: 'update',
              inputDiff: true,
            },
          },
          old_inputs: {
            resourceName: 'my-aks',
            kubernetesVersion: '1.33',
            location: 'centralindia',
            agentPoolProfiles: [
              {
                name: 'agents',
                count: 1,
                vmSize: 'Standard_D2s_v3',
                availabilityZones: null,
              },
            ],
            identity: {
              type: 'SystemAssigned',
              userAssignedIdentities: null,
            },
            tags: null,
          },
          new_inputs: {
            resourceName: 'my-aks',
            kubernetesVersion: '1.33',
            location: 'centralindia',
            agentPoolProfiles: [
              {
                name: 'agents',
                count: 1,
                vmSize: 'Standard_D2s_v3',
                availabilityZones: [],
              },
            ],
            identity: {
              type: 'SystemAssigned',
              userAssignedIdentities: {},
            },
            tags: {},
          },
        },
        {
          urn: 'urn:pulumi:prod::infra::azure-native:network:VirtualNetwork::vnet',
          type: 'azure-native:network:VirtualNetwork',
          name: 'vnet',
          action: 'same',
        },
        {
          urn: 'urn:pulumi:prod::infra::azure-native:network:Subnet::private-subnet',
          type: 'azure-native:network:Subnet',
          name: 'private-subnet',
          action: 'same',
        },
        {
          urn: 'urn:pulumi:prod::infra::azure-native:compute:SshPublicKey::ssh-key',
          type: 'azure-native:compute:SshPublicKey',
          name: 'ssh-key',
          action: 'same',
        },
        {
          urn: 'urn:pulumi:prod::infra::azure-native:authorization:RoleAssignment::acr-pull',
          type: 'azure-native:authorization:RoleAssignment',
          name: 'acr-pull',
          action: 'same',
        },
        {
          urn: 'urn:pulumi:prod::infra::random:index:RandomPet::ssh-key-name',
          type: 'random:index:RandomPet',
          name: 'ssh-key-name',
          action: 'same',
        },
      ],
    }}
  />
)

export const RBACArrayNoise = () => (
  <PulumiDiff
    plan={{
      stdout: '',
      stderr: '',
      change_summary: { update: 1 },
      resource_changes: [
        {
          urn: 'urn:pulumi:prod::k8s-platform::kubernetes:rbac.authorization.k8s.io/v1:ClusterRole::system-node',
          type: 'kubernetes:rbac.authorization.k8s.io/v1:ClusterRole',
          name: 'system-node',
          action: 'update',
          diffs: ['rules'],
          detailed_diff: {
            'rules[4].verbs[3]': { kind: 'add', inputDiff: true },
          },
          old_inputs: {
            metadata: { name: 'system:node' },
            rules: [
              {
                apiGroups: [''],
                resources: ['nodes'],
                verbs: ['get', 'list', 'watch'],
                resourceNames: null,
              },
              {
                apiGroups: [''],
                resources: ['pods'],
                verbs: ['get', 'list', 'watch'],
                resourceNames: null,
              },
              {
                apiGroups: [''],
                resources: ['services'],
                verbs: ['get', 'list'],
                resourceNames: null,
              },
              {
                apiGroups: ['apps'],
                resources: ['daemonsets'],
                verbs: ['get', 'list', 'watch'],
                resourceNames: null,
              },
              {
                apiGroups: ['coordination.k8s.io'],
                resources: ['leases'],
                verbs: ['get', 'create', 'update'],
                resourceNames: null,
              },
            ],
          },
          new_inputs: {
            metadata: { name: 'system:node' },
            rules: [
              {
                apiGroups: [''],
                resources: ['nodes'],
                verbs: ['get', 'list', 'watch'],
                resourceNames: [],
              },
              {
                apiGroups: [''],
                resources: ['pods'],
                verbs: ['get', 'list', 'watch'],
                resourceNames: [],
              },
              {
                apiGroups: [''],
                resources: ['services'],
                verbs: ['get', 'list'],
                resourceNames: [],
              },
              {
                apiGroups: ['apps'],
                resources: ['daemonsets'],
                verbs: ['get', 'list', 'watch'],
                resourceNames: [],
              },
              {
                apiGroups: ['coordination.k8s.io'],
                resources: ['leases'],
                verbs: ['get', 'create', 'update', 'patch'],
                resourceNames: [],
              },
            ],
          },
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
