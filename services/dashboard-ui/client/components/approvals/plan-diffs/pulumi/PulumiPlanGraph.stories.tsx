export default {
  title: 'Approvals/PlanDiffs/PulumiPlanGraph',
}

import { PulumiPlanGraph } from './PulumiPlanGraph'

export const AllActions = () => (
  <PulumiPlanGraph
    resources={[
      { urn: 'urn:pulumi:prod::app::aws:s3/bucket:Bucket::new-bucket', type: 'aws:s3/bucket:Bucket', name: 'new-bucket', action: 'create' },
      { urn: 'urn:pulumi:prod::app::aws:ec2/securityGroup:SecurityGroup::api-sg', type: 'aws:ec2/securityGroup:SecurityGroup', name: 'api-sg', action: 'update' },
      { urn: 'urn:pulumi:prod::app::aws:cloudwatch/metricAlarm:MetricAlarm::old-alarm', type: 'aws:cloudwatch/metricAlarm:MetricAlarm', name: 'old-alarm', action: 'delete' },
      { urn: 'urn:pulumi:prod::app::aws:rds/instance:Instance::primary-db', type: 'aws:rds/instance:Instance', name: 'primary-db', action: 'replace' },
      { urn: 'urn:pulumi:prod::app::aws:rds/instance:Instance::primary-db-new', type: 'aws:rds/instance:Instance', name: 'primary-db-new', action: 'create-replacement' },
      { urn: 'urn:pulumi:prod::app::aws:rds/instance:Instance::primary-db-old', type: 'aws:rds/instance:Instance', name: 'primary-db-old', action: 'delete-replaced' },
      { urn: 'urn:pulumi:prod::app::aws:ec2/getSubnets:getSubnets::subnets', type: 'aws:ec2/getSubnets:getSubnets', name: 'subnets', action: 'read' },
      { urn: 'urn:pulumi:prod::app::aws:ec2/getAmi:getAmi::latest-ami', type: 'aws:ec2/getAmi:getAmi', name: 'latest-ami', action: 'refresh' },
      { urn: 'urn:pulumi:prod::app::aws:ec2/vpc:Vpc::main-vpc', type: 'aws:ec2/vpc:Vpc', name: 'main-vpc', action: 'same' },
      { urn: 'urn:pulumi:prod::app::aws:iam/role:Role::task-role', type: 'aws:iam/role:Role', name: 'task-role', action: 'same' },
    ]}
  />
)

export const SimpleCreates = () => (
  <PulumiPlanGraph
    resources={[
      { urn: 'urn:pulumi:prod::data::aws:s3/bucket:Bucket::artifacts', type: 'aws:s3/bucket:Bucket', name: 'artifacts', action: 'create' },
      { urn: 'urn:pulumi:prod::data::aws:s3/bucketPolicy:BucketPolicy::artifacts-policy', type: 'aws:s3/bucketPolicy:BucketPolicy', name: 'artifacts-policy', action: 'create' },
      { urn: 'urn:pulumi:prod::data::aws:s3/bucketNotification:BucketNotification::artifacts-notify', type: 'aws:s3/bucketNotification:BucketNotification', name: 'artifacts-notify', action: 'create' },
    ]}
  />
)

export const MixedActions = () => (
  <PulumiPlanGraph
    resources={[
      { urn: 'urn:pulumi:prod::net::aws:ec2/vpc:Vpc::main-vpc', type: 'aws:ec2/vpc:Vpc', name: 'main-vpc', action: 'same' },
      { urn: 'urn:pulumi:prod::net::aws:ec2/securityGroup:SecurityGroup::api-sg', type: 'aws:ec2/securityGroup:SecurityGroup', name: 'api-sg', action: 'update' },
      { urn: 'urn:pulumi:prod::compute::aws:lambda/function:Function::processor', type: 'aws:lambda/function:Function', name: 'processor', action: 'create' },
      { urn: 'urn:pulumi:prod::mon::aws:cloudwatch/metricAlarm:MetricAlarm::legacy', type: 'aws:cloudwatch/metricAlarm:MetricAlarm', name: 'legacy', action: 'delete' },
      { urn: 'urn:pulumi:prod::net::aws:ec2/getSubnets:getSubnets::private', type: 'aws:ec2/getSubnets:getSubnets', name: 'private', action: 'read' },
    ]}
  />
)

export const DatabaseReplace = () => (
  <PulumiPlanGraph
    resources={[
      { urn: 'urn:pulumi:prod::db::aws:rds/instance:Instance::primary', type: 'aws:rds/instance:Instance', name: 'primary', action: 'replace' },
      { urn: 'urn:pulumi:prod::db::aws:rds/instance:Instance::primary-new', type: 'aws:rds/instance:Instance', name: 'primary-new', action: 'create-replacement' },
      { urn: 'urn:pulumi:prod::db::aws:rds/instance:Instance::primary-old', type: 'aws:rds/instance:Instance', name: 'primary-old', action: 'delete-replaced' },
    ]}
  />
)

export const KubernetesResources = () => (
  <PulumiPlanGraph
    resources={[
      { urn: 'urn:pulumi:prod::app::kubernetes:core/v1:Namespace::app-ns', type: 'kubernetes:core/v1:Namespace', name: 'app-ns', action: 'same' },
      { urn: 'urn:pulumi:prod::app::kubernetes:apps/v1:Deployment::api', type: 'kubernetes:apps/v1:Deployment', name: 'api', action: 'update' },
      { urn: 'urn:pulumi:prod::app::kubernetes:apps/v1:Deployment::worker', type: 'kubernetes:apps/v1:Deployment', name: 'worker', action: 'update' },
      { urn: 'urn:pulumi:prod::app::kubernetes:core/v1:Service::api-svc', type: 'kubernetes:core/v1:Service', name: 'api-svc', action: 'same' },
      { urn: 'urn:pulumi:prod::app::kubernetes:core/v1:ConfigMap::config', type: 'kubernetes:core/v1:ConfigMap', name: 'config', action: 'update' },
      { urn: 'urn:pulumi:prod::app::kubernetes:networking.k8s.io/v1:NetworkPolicy::netpol', type: 'kubernetes:networking.k8s.io/v1:NetworkPolicy', name: 'netpol', action: 'create' },
      { urn: 'urn:pulumi:prod::app::kubernetes:policy/v1:PodDisruptionBudget::pdb', type: 'kubernetes:policy/v1:PodDisruptionBudget', name: 'pdb', action: 'create' },
      { urn: 'urn:pulumi:prod::app::kubernetes:autoscaling/v2:HorizontalPodAutoscaler::hpa', type: 'kubernetes:autoscaling/v2:HorizontalPodAutoscaler', name: 'hpa', action: 'update' },
      { urn: 'urn:pulumi:prod::app::kubernetes:helm.sh/v4:Chart::monitoring', type: 'kubernetes:helm.sh/v4:Chart', name: 'monitoring', action: 'create' },
    ]}
  />
)

export const LargePlan = () => {
  const types = [
    'aws:ec2/instance:Instance',
    'aws:ec2/securityGroup:SecurityGroup',
    'aws:s3/bucket:Bucket',
    'aws:iam/role:Role',
    'aws:lambda/function:Function',
    'aws:sqs/queue:Queue',
    'kubernetes:apps/v1:Deployment',
    'kubernetes:core/v1:Service',
  ]
  const actions = ['create', 'update', 'delete', 'same', 'replace', 'read']

  const resources = types.flatMap((type, ti) =>
    Array.from({ length: 3 }, (_, i) => ({
      urn: `urn:pulumi:prod::app::${type}::item-${ti}-${i}`,
      type,
      name: `item-${ti}-${i}`,
      action: actions[(ti + i) % actions.length],
    }))
  )

  return <PulumiPlanGraph resources={resources} />
}

export const AzureResources = () => (
  <PulumiPlanGraph
    resources={[
      { urn: 'urn:pulumi:prod::infra::azure-native:containerregistry:Registry::acr', type: 'azure-native:containerregistry:Registry', name: 'acr', action: 'update' },
      { urn: 'urn:pulumi:prod::infra::azure-native:containerservice:ManagedCluster::aks', type: 'azure-native:containerservice:ManagedCluster', name: 'aks', action: 'update' },
      { urn: 'urn:pulumi:prod::infra::azure-native:network:VirtualNetwork::vnet', type: 'azure-native:network:VirtualNetwork', name: 'vnet', action: 'same' },
      { urn: 'urn:pulumi:prod::infra::azure-native:network:Zone::public-zone', type: 'azure-native:network:Zone', name: 'public-zone', action: 'update' },
      { urn: 'urn:pulumi:prod::infra::azure-native:network:Subnet::private', type: 'azure-native:network:Subnet', name: 'private', action: 'same' },
    ]}
  />
)

export const Empty = () => <PulumiPlanGraph resources={[]} />
